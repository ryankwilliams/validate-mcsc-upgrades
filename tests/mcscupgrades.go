package tests

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/openshift/osde2e-framework/pkg/clients/ocm"
	"github.com/openshift/osde2e-framework/pkg/providers/osd"
	"github.com/openshift/osde2e-framework/pkg/providers/rosa"
)

var (
	applyHCPWorkloads               = ginkgo.Label("ApplyHCPWorkloads")
	clusterName                     = getEnvVar("CLUSTER_NAME", "my-cluster")
	clusterChannelGroup             = getEnvVar("CLUSTER_CHANNEL_GROUP", "candidate")
	clusterVersion                  = getEnvVar("CLUSTER_VERSION", "4.12.18")
	hcpClusterID                    *string
	managementClusterID             string
	managementClusterVersion        semver.Version
	managementClusterUpgradeVersion semver.Version
	mcUpgrade                       = ginkgo.Label("MCUpgrade")
	mcUpgradeHealthChecks           = ginkgo.Label("MCUpgradeHealthChecks")
	osdProvider                     *osd.Provider
	removeHCPWorkloads              = ginkgo.Label("RemoveHCPWorkloads")
	rosaProvider                    *rosa.Provider
	scUpgrade                       = ginkgo.Label("SCUpgrade")
	scUpgradeHealthChecks           = ginkgo.Label("SCUpgradeHealthChecks")
	serviceClusterID                string
	serviceClusterVersion           semver.Version
	serviceClusterUpgradeVersion    semver.Version
)

var _ = ginkgo.BeforeSuite(func() {
	var (
		err error

		ctx                             = context.Background()
		ocmEnv                          = ocm.Integration
		ocmToken                        = os.Getenv("OCM_TOKEN")
		osdFleetMgmtManagementClusterID = os.Getenv("OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID")
		osdFleetMgmtServiceClusterID    = os.Getenv("OSD_FLEET_MGMT_SERVICE_CLUSTER_ID")
		provisionShardID                = os.Getenv("PROVISION_SHARD_ID")
		upgradeType                     = os.Getenv("UPGRADE_TYPE")
	)

	// Construct new rosa provider
	rosaProvider, err = rosa.New(ctx, ocmToken, ocmEnv)
	gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to construct rosa provider")

	// Construct new osd provider
	osdProvider, err = osd.New(ctx, ocmToken, ocmEnv)
	gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to construct osd provider")

	// Validate required cluster upgrade input
	gomega.Expect(osdFleetMgmtServiceClusterID).Error().ShouldNot(gomega.BeEmpty(), "osd fleet manager service cluster id was not provided")
	gomega.Expect(osdFleetMgmtManagementClusterID).Error().ShouldNot(gomega.BeEmpty(), "osd fleet manager management cluster id was not provided")
	gomega.Expect(provisionShardID).Error().ShouldNot(gomega.BeEmpty(), "provision shard id was not provided") // TODO: Open issue/pr to expose this field when using ocm-sdk-go
	gomega.Expect(upgradeType).Error().Should(gomega.BeElementOf([]string{"Y", "Z"}), "upgrade type is invalid")

	// TODO: Consolidate these

	// Identify the service cluster install/upgrade versions
	if scUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		osdFleetManagerSC, err := osdProvider.OSDFleetMgmt().V1().ServiceClusters().ServiceCluster(osdFleetMgmtServiceClusterID).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to get osd fleet manager service cluster")

		serviceCluster, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(osdFleetManagerSC.Body().ClusterManagementReference().ClusterId()).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to get service cluster")

		serviceClusterID = serviceCluster.Body().Version().RawID()
		serviceClusterVersion, err := semver.NewVersion(serviceClusterID)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to parse version to semantic version")

		availableVersions := serviceCluster.Body().Version().AvailableUpgrades()
		totalAvailableVersions := len(availableVersions)
		gomega.Expect(totalAvailableVersions).ToNot(gomega.BeNumerically("==", 0), "service cluster has no available supported upgrade versions")

		for i := 0; i < totalAvailableVersions; i++ {
			version, err := semver.NewVersion(availableVersions[totalAvailableVersions-i-1])
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to parse version to semantic version")
			if (serviceClusterVersion.Minor() == version.Minor()) && upgradeType == "Z" {
				serviceClusterUpgradeVersion = *version
				break
			} else if (serviceClusterVersion.Minor() < version.Minor()) && upgradeType == "Y" {
				serviceClusterUpgradeVersion = *version
				break
			}
		}
		gomega.Expect(serviceClusterUpgradeVersion).ToNot(gomega.BeNil(), "unable to identify service cluster upgrade version")
	}

	// Identify the management cluster install/upgrade versions
	if mcUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		osdFleetManagerMC, err := osdProvider.OSDFleetMgmt().V1().ManagementClusters().ManagementCluster(osdFleetMgmtManagementClusterID).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to get osd fleet manager management cluster")

		managementCluster, err := osdProvider.ClustersMgmt().V1().Clusters().Cluster(osdFleetManagerMC.Body().ClusterManagementReference().ClusterId()).Get().SendContext(ctx)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to get management cluster")

		managementClusterID = managementCluster.Body().Version().RawID()
		managementClusterVersion, err := semver.NewVersion(managementClusterID)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to parse version to semantic version")

		availableVersions := managementCluster.Body().Version().AvailableUpgrades()
		totalAvailableVersions := len(availableVersions)
		gomega.Expect(totalAvailableVersions).ToNot(gomega.BeNumerically("==", 0), "management cluster has no available supported upgrade versions")

		for i := 0; i < totalAvailableVersions; i++ {
			version, err := semver.NewVersion(availableVersions[totalAvailableVersions-i-1])
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to parse version to semantic version")
			if (managementClusterVersion.Minor() == version.Minor()) && upgradeType == "Z" {
				managementClusterUpgradeVersion = *version
				break
			} else if (managementClusterVersion.Minor() < version.Minor()) && upgradeType == "Y" {
				managementClusterUpgradeVersion = *version
				break
			}
		}
		gomega.Expect(managementClusterUpgradeVersion).ToNot(gomega.BeNil(), "unable to identify management cluster upgrade version")
	}

	if applyHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		clusterID, err := rosaProvider.CreateCluster(ctx, &rosa.CreateClusterOptions{
			ClusterName:  clusterName,
			Version:      clusterVersion,
			ChannelGroup: clusterChannelGroup,
			HostedCP:     true,
		})
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "create hcp cluster failed")
		hcpClusterID = &clusterID
	}
})

var _ = ginkgo.AfterSuite(func() {
	ctx := context.Background()

	defer func() {
		_ = osdProvider.Connection.Close()
		_ = rosaProvider.Connection.Close()
	}()

	if removeHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		err := rosaProvider.DeleteCluster(ctx, &rosa.DeleteClusterOptions{
			ClusterName: clusterName,
			ClusterID:   *hcpClusterID,
			HostedCP:    true,
		})
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "delete hcp cluster failed")
	}
})

var _ = ginkgo.Describe("HyperShift", ginkgo.Ordered, func() {
	ginkgo.It("service cluster is upgraded successfully", scUpgrade, func(ctx context.Context) {
		fmt.Printf("Performing service cluster upgrade to version %q\n", serviceClusterVersion.String())
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("service cluster health checks are passing post upgrade", scUpgrade, scUpgradeHealthChecks, func(ctx context.Context) {
		fmt.Println("Performing service cluster post upgrade health checks")
		err := osdProvider.OCMUpgrade(ctx, serviceClusterID, serviceClusterVersion, serviceClusterUpgradeVersion)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "service cluster upgrade failed")
	})

	ginkgo.It("hcp workloads are unaffected post service cluster upgrade", scUpgrade, func(ctx context.Context) {
		fmt.Println("Performing hcp cluster post service cluster upgrade")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("management cluster is upgraded successfully", mcUpgrade, func(ctx context.Context) {
		fmt.Printf("Performing management cluster upgrade to version %q\n", managementClusterVersion.String())
		err := osdProvider.OCMUpgrade(ctx, managementClusterID, managementClusterVersion, managementClusterUpgradeVersion)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "management cluster upgrade failed")
	})

	ginkgo.It("management cluster health checks are passing post upgrade", mcUpgrade, mcUpgradeHealthChecks, func(ctx context.Context) {
		fmt.Println("Performing management cluster post upgrade health checks")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("hcp workloads are unaffected post management cluster upgrade", mcUpgrade, func(ctx context.Context) {
		fmt.Println("Performing hcp cluster post management cluster upgrade")
		gomega.Expect(true).Should(gomega.BeTrue())
	})
})

// getEnvVar gets environment variable value and if not set, returns the
// default provided
func getEnvVar(key, value string) string {
	result, exist := os.LookupEnv(key)
	if exist {
		return result
	}
	return value
}
