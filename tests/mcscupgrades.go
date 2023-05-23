package tests

import (
	"context"
	"os"

	"github.com/Masterminds/semver"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	kubernetesclient "github.com/openshift/osde2e-framework/pkg/clients/kubernetes"
	"github.com/openshift/osde2e-framework/pkg/clients/ocm"
	"github.com/openshift/osde2e-framework/pkg/providers/osd"
	"github.com/openshift/osde2e-framework/pkg/providers/rosa"

	"k8s.io/utils/pointer"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	applyHCPWorkloads               = ginkgo.Label("ApplyHCPWorkloads")
	clusterName                     = getEnvVar("CLUSTER_NAME", envconf.RandomName("hcp", 4))
	clusterChannelGroup             = getEnvVar("CLUSTER_CHANNEL_GROUP", "candidate")
	clusterVersion                  = getEnvVar("CLUSTER_VERSION", "4.12.18")
	hcpClusterKubeConfigFile        *string
	hcpClusterID                    *string
	managementClusterKubeConfigFile *string
	managementClusterID             *string
	managementClusterVersion        *semver.Version
	managementClusterUpgradeVersion *semver.Version
	mcUpgrade                       = ginkgo.Label("MCUpgrade")
	mcUpgradeHealthChecks           = ginkgo.Label("MCUpgradeHealthChecks")
	osdProvider                     *osd.Provider
	removeHCPWorkloads              = ginkgo.Label("RemoveHCPWorkloads")
	rosaProvider                    *rosa.Provider
	scUpgrade                       = ginkgo.Label("SCUpgrade")
	scUpgradeHealthChecks           = ginkgo.Label("SCUpgradeHealthChecks")
	serviceClusterKubeConfigFile    *string
	serviceClusterID                *string
	serviceClusterVersion           *semver.Version
	serviceClusterUpgradeVersion    *semver.Version
)

var _ = ginkgo.BeforeSuite(func() {
	var (
		ctx                             = context.Background()
		ocmEnv                          = ocm.Integration
		ocmToken                        = os.Getenv("OCM_TOKEN")
		osdFleetMgmtManagementClusterID = os.Getenv("OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID")
		osdFleetMgmtServiceClusterID    = os.Getenv("OSD_FLEET_MGMT_SERVICE_CLUSTER_ID")
		provisionShardID                = os.Getenv("PROVISION_SHARD_ID")
		upgradeType                     = os.Getenv("UPGRADE_TYPE")
	)

	// Validate required general input
	gomega.Expect(ocmToken).Error().ShouldNot(gomega.BeEmpty(), "ocm token is undefined")

	// Construct new rosa provider
	rosaProvider, err := rosa.New(ctx, ocmToken, ocmEnv)
	gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to construct rosa provider")

	// Construct new osd provider
	osdProvider, err = osd.New(ctx, ocmToken, ocmEnv)
	gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to construct osd provider")

	// Validate required cluster upgrade input
	gomega.Expect(osdFleetMgmtServiceClusterID).Error().ShouldNot(gomega.BeEmpty(), "osd fleet manager service cluster id is undefined")
	gomega.Expect(osdFleetMgmtManagementClusterID).Error().ShouldNot(gomega.BeEmpty(), "osd fleet manager management cluster id is undefined")
	gomega.Expect(provisionShardID).Error().ShouldNot(gomega.BeEmpty(), "provision shard id is undefined")
	gomega.Expect(upgradeType).Error().Should(gomega.BeElementOf([]string{"Y", "Z"}), "upgrade type is invalid")

	// Identify the service cluster install/upgrade versions
	if scUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		osdFleetManagerSC, err := osdProvider.OSDFleetMgmt().V1().ServiceClusters().ServiceCluster(osdFleetMgmtServiceClusterID).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "ocm osd fleet manager api request failed to get osd fleet manager service cluster: %q", osdFleetMgmtServiceClusterID)
		gomega.Expect(osdFleetManagerSC).NotTo(gomega.BeNil(), "osd fleet manager service cluster: %q does not exist", osdFleetMgmtServiceClusterID)

		serviceCluster, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(osdFleetManagerSC.Body().ClusterManagementReference().ClusterId()).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to get get service cluster: %s", osdFleetMgmtServiceClusterID)

		serviceClusterID = pointer.String(serviceCluster.Body().ID())
		serviceClusterVersion, err := semver.NewVersion(serviceCluster.Body().Version().RawID())
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to parse service cluster installed version to semantic version")

		availableVersions := serviceCluster.Body().Version().AvailableUpgrades()
		totalUpgradeVersionsAvailable := len(availableVersions)
		gomega.Expect(totalUpgradeVersionsAvailable).ToNot(gomega.BeNumerically("==", 0), "service cluster has no available supported upgrade versions")

		for i := 0; i < totalUpgradeVersionsAvailable; i++ {
			version, err := semver.NewVersion(availableVersions[totalUpgradeVersionsAvailable-i-1])
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to parse service cluster upgrade version to semantic version")
			if (serviceClusterVersion.Minor() == version.Minor()) && upgradeType == "Z" {
				serviceClusterUpgradeVersion = version
				break
			} else if (serviceClusterVersion.Minor() < version.Minor()) && upgradeType == "Y" {
				serviceClusterUpgradeVersion = version
				break
			}
		}
		gomega.Expect(serviceClusterUpgradeVersion).ToNot(gomega.BeNil(), "failed to identify service cluster %q upgrade version", osdFleetMgmtServiceClusterID)

		// Get kubeconfig
		kubeConfigFile, err := osdProvider.KubeConfigFile(ctx, *serviceClusterID)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to get service cluster %q kubeconfig file", osdFleetMgmtServiceClusterID)
		serviceClusterKubeConfigFile = &kubeConfigFile
	}

	// Identify the management cluster install/upgrade versions
	if mcUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		osdFleetManagerMC, err := osdProvider.OSDFleetMgmt().V1().ManagementClusters().ManagementCluster(osdFleetMgmtManagementClusterID).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "ocm osd fleet manager api request failed to get osd fleet manager management cluster: %q", osdFleetMgmtManagementClusterID)
		gomega.Expect(osdFleetManagerMC).NotTo(gomega.BeNil(), "osd fleet manager management cluster: %q does not exist", osdFleetMgmtManagementClusterID)

		managementCluster, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(osdFleetManagerMC.Body().ClusterManagementReference().ClusterId()).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to get get management cluster: %s", osdFleetMgmtManagementClusterID)

		managementClusterID = pointer.String(managementCluster.Body().ID())
		managementClusterVersion, err = semver.NewVersion(managementCluster.Body().Version().RawID())
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to parse management cluster installed version to semantic version")

		availableVersions := managementCluster.Body().Version().AvailableUpgrades()
		totalAvailableVersions := len(availableVersions)
		gomega.Expect(totalAvailableVersions).ToNot(gomega.BeNumerically("==", 0), "management cluster has no available supported upgrade versions")

		for i := 0; i < totalAvailableVersions; i++ {
			version, err := semver.NewVersion(availableVersions[totalAvailableVersions-i-1])
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to parse management cluster upgrade version to semantic version")
			if (managementClusterVersion.Minor() == version.Minor()) && upgradeType == "Z" {
				managementClusterUpgradeVersion = version
				break
			} else if (managementClusterVersion.Minor() < version.Minor()) && upgradeType == "Y" {
				managementClusterUpgradeVersion = version
				break
			}
		}
		gomega.Expect(managementClusterUpgradeVersion).ToNot(gomega.BeNil(), "failed to identify service cluster %q upgrade version", osdFleetMgmtManagementClusterID)

		// Get kubeconfig
		kubeConfigFile, err := osdProvider.KubeConfigFile(ctx, *managementClusterID)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to get management cluster %q kubeconfig file", osdFleetMgmtServiceClusterID)
		managementClusterKubeConfigFile = &kubeConfigFile
	}

	if applyHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		clusterID, err := rosaProvider.CreateCluster(ctx, &rosa.CreateClusterOptions{
			ClusterName:  clusterName,
			Version:      clusterVersion,
			ChannelGroup: clusterChannelGroup,
			HostedCP:     true,
		})
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to create rosa hosted control plane cluster")
		hcpClusterID = &clusterID

		kubeConfigFile, err := rosaProvider.KubeConfigFile(ctx, *hcpClusterID)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to get rosa hosted control plane cluster %q kubeconfig file", hcpClusterID)
		hcpClusterKubeConfigFile = &kubeConfigFile
	}
})

var _ = ginkgo.AfterSuite(func() {
	ctx := context.Background()

	defer func() {
		_ = osdProvider.Client.Close()
	}()

	if removeHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		err := rosaProvider.DeleteCluster(ctx, &rosa.DeleteClusterOptions{
			ClusterName: clusterName,
			ClusterID:   *hcpClusterID,
			HostedCP:    true,
		})
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to delete rosa hosted control plane cluster")
	}
})

var _ = ginkgo.Describe("HyperShift", ginkgo.Ordered, func() {
	kubernetesClient := func(kubeconfigFile string) (*kubernetesclient.Client, error) {
		os.Setenv("KUBECONFIG", kubeconfigFile)
		return kubernetesclient.New()
	}

	hcpClusterCheck := func() error {
		_, err := kubernetesClient(*hcpClusterKubeConfigFile)
		return err
	}

	ginkgo.It("service cluster is upgraded successfully", scUpgrade, func(ctx context.Context) {
		client, err := kubernetesClient(*serviceClusterKubeConfigFile)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to construct kubernetes client to service cluster")

		err = osdProvider.OCMUpgrade(ctx, client, *serviceClusterID, *serviceClusterVersion, *serviceClusterUpgradeVersion)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "service cluster upgrade failed")
	})

	ginkgo.It("service cluster health checks are passing post upgrade", scUpgrade, scUpgradeHealthChecks, func(ctx context.Context) {
		_, err := kubernetesClient(*serviceClusterKubeConfigFile)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to construct kubernetes client to service cluster")
	})

	ginkgo.It("hcp workloads are unaffected post service cluster upgrade", scUpgrade, func(ctx context.Context) {
		err := hcpClusterCheck()
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "hosted control plane cluster failed post upgrade check")
	})

	ginkgo.It("management cluster is upgraded successfully", mcUpgrade, func(ctx context.Context) {
		client, err := kubernetesClient(*managementClusterKubeConfigFile)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to construct kubernetes client to management cluster")

		err = osdProvider.OCMUpgrade(ctx, client, *managementClusterID, *managementClusterVersion, *managementClusterUpgradeVersion)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "management cluster upgrade failed")
	})

	ginkgo.It("management cluster health checks are passing post upgrade", mcUpgrade, mcUpgradeHealthChecks, func(ctx context.Context) {
		_, err := kubernetesClient(*managementClusterKubeConfigFile)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "failed to construct kubernetes client to management cluster")
	})

	ginkgo.It("hcp workloads are unaffected post management cluster upgrade", mcUpgrade, func(ctx context.Context) {
		err := hcpClusterCheck()
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "hosted control plane cluster failed post upgrade check")
	})
})

// getEnvVar gets environment variable value and returns default if unset
func getEnvVar(key, value string) string {
	result, exist := os.LookupEnv(key)
	if exist {
		return result
	}
	return value
}
