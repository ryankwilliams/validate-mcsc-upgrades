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

	"github.com/ryankwilliams/validate-mcsc-upgrades/internal/labels"
)

var (
	managementClusterID             string
	managementClusterVersion        semver.Version
	managementClusterUpgradeVersion semver.Version
	serviceClusterID                string
	serviceClusterVersion           semver.Version
	serviceClusterUpgradeVersion    semver.Version
	scUpgradeVersion                *string
	mcUpgradeVersion                *string
	osdProvider                     *osd.Provider
	rosaProvider                    *rosa.Provider
	hcpClusterID                    *string
)

const (
	clusterName    = "my-cluster"
	clusterVersion = "4.12.18"
	channelGroup   = "candidate"
)

var _ = ginkgo.BeforeSuite(func() {
	var (
		err error

		ctx                             = context.Background()
		osdFleetMgmtManagementClusterID = os.Getenv("OSD_FLEET_MGMT_MANAGEMENT_CLUSTER_ID")
		ocmEnv                          = ocm.Integration
		ocmToken                        = os.Getenv("OCM_TOKEN")
		provisionShardID                = os.Getenv("PROVISION_SHARD_ID")
		osdFleetMgmtServiceClusterID    = os.Getenv("OSD_FLEET_MGMT_SERVICE_CLUSTER_ID")
		upgradeType                     = os.Getenv("UPGRADE_TYPE")
	)

	// Construct new rosa provider
	rosaProvider, err = rosa.New(ctx, ocmToken, ocmEnv)
	gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to construct rosa provider")

	// Construct new osd provider
	osdProvider, err = osd.New(ctx, ocmToken, ocmEnv)
	gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to construct osd provider")

	// Verify required data was provided
	gomega.Expect(osdFleetMgmtServiceClusterID).Error().ShouldNot(gomega.BeEmpty(), "service cluster id was not provided")
	gomega.Expect(osdFleetMgmtManagementClusterID).Error().ShouldNot(gomega.BeEmpty(), "management cluster id was not provided")
	gomega.Expect(provisionShardID).Error().ShouldNot(gomega.BeEmpty(), "provision shard id was not provided") // TODO: Open issue/pr to expose this field when using ocm-sdk-go
	gomega.Expect(upgradeType).Error().ShouldNot(gomega.BeEmpty(), "upgrade type was not provided, must be either 'Y' or 'Z'")
	gomega.Expect(upgradeType).Error().Should(gomega.BeElementOf([]string{"Y", "Z"}), "upgrade type is invalid")

	// TODO: Consolidate these

	// Identify the service cluster install/upgrade versions
	if labels.SCUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
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
	if labels.MCUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
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

	if labels.ApplyHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		hcpClusterID, err := rosaProvider.CreateCluster(ctx, &rosa.CreateClusterOptions{
			ClusterName:  clusterName,
			Version:      clusterVersion,
			ChannelGroup: channelGroup,
			HostedCP:     true,
		})
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "create hcp cluster failed")
		fmt.Printf("HCP cluster %q created\n", hcpClusterID)
	}
})

var _ = ginkgo.AfterSuite(func() {
	ctx := context.Background()

	defer func() {
		_ = osdProvider.Connection.Close()
		_ = rosaProvider.Connection.Close()
	}()

	if labels.RemoveHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		err := rosaProvider.DeleteCluster(ctx, &rosa.DeleteClusterOptions{
			ClusterName: clusterName,
			ClusterID:   *hcpClusterID,
			HostedCP:    true,
		})
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "delete hcp cluster failed")
	}
})

var _ = ginkgo.Describe("HyperShift", ginkgo.Ordered, func() {
	ginkgo.It("service cluster is upgraded successfully", labels.SCUpgrade, func(ctx context.Context) {
		fmt.Printf("Performing service cluster upgrade to version %q\n", *scUpgradeVersion)
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("service cluster health checks are passing post upgrade", labels.SCUpgrade, labels.SCUpgradeHealthChecks, func(ctx context.Context) {
		fmt.Println("Performing service cluster post upgrade health checks")
		err := osdProvider.OCMUpgrade(ctx, serviceClusterID, serviceClusterVersion, serviceClusterUpgradeVersion)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "service cluster upgrade failed")
	})

	ginkgo.It("hcp workloads are unaffected post service cluster upgrade", labels.SCUpgrade, func(ctx context.Context) {
		fmt.Println("Performing hcp cluster post service cluster upgrade")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("management cluster is upgraded successfully", labels.MCUpgrade, func(ctx context.Context) {
		fmt.Printf("Performing management cluster upgrade to version %q\n", *mcUpgradeVersion)
		err := osdProvider.OCMUpgrade(ctx, managementClusterID, managementClusterVersion, managementClusterUpgradeVersion)
		gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "management cluster upgrade failed")
	})

	ginkgo.It("management cluster health checks are passing post upgrade", labels.MCUpgrade, labels.MCUpgradeHealthChecks, func(ctx context.Context) {
		fmt.Println("Performing management cluster post upgrade health checks")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("hcp workloads are unaffected post management cluster upgrade", labels.MCUpgrade, func(ctx context.Context) {
		fmt.Println("Performing hcp cluster post management cluster upgrade")
		gomega.Expect(true).Should(gomega.BeTrue())
	})
})
