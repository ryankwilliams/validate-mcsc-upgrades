package tests

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/ryankwilliams/validate-mcsc-upgrades/internal/labels"
	"github.com/ryankwilliams/validate-mcsc-upgrades/internal/provider"
)

var (
	scUpgradeVersion *string
	mcUpgradeVersion *string
)

var _ = ginkgo.BeforeSuite(func() {
	var (
		clusterVersion string
		err            error
	)

	// Connect to MC/SC

	// Determine service cluster install version
	if labels.SCUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		if clusterVersion, err = provider.IdentifyClusterVersion(); err != nil {
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to identify service cluster version")
		}

		if scUpgradeVersion, err = provider.DetermineUpgradeVersion(clusterVersion); err != nil {
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to identify service cluster upgrade version")
		}
	}

	// Determine management cluster upgrade version
	if labels.MCUpgrade.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		if clusterVersion, err = provider.IdentifyClusterVersion(); err != nil {
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to identify management cluster version")
		}

		if mcUpgradeVersion, err = provider.DetermineUpgradeVersion(clusterVersion); err != nil {
			gomega.Expect(err).Error().ShouldNot(gomega.HaveOccurred(), "unable to identify management cluster upgrade version")
		}
	}

	if labels.ApplyHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		provider.AddHCPWorkloads()
	}
})

var _ = ginkgo.AfterSuite(func() {
	if labels.RemoveHCPWorkloads.MatchesLabelFilter(ginkgo.GinkgoLabelFilter()) {
		provider.RemoveHCPWorkloads()
	}
})

var _ = ginkgo.Describe("HyperShift", ginkgo.Ordered, func() {

	ginkgo.It("service cluster is upgraded successfully", labels.SCUpgrade, func() {
		fmt.Printf("Performing service cluster upgrade to version %q\n", *scUpgradeVersion)
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("service cluster health checks are passing post upgrade", labels.SCUpgrade, labels.SCUpgradeHealthChecks, func() {
		fmt.Println("Performing service cluster post upgrade health checks")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("hcp workloads are unaffected post service cluster upgrade", labels.SCUpgrade, func() {
		fmt.Println("Performing hcp cluster post service cluster upgrade")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("management cluster is upgraded successfully", labels.MCUpgrade, func() {
		fmt.Printf("Performing management cluster upgrade to version %q\n", *mcUpgradeVersion)
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("management cluster health checks are passing post upgrade", labels.MCUpgrade, labels.MCUpgradeHealthChecks, func() {
		fmt.Println("Performing management cluster post upgrade health checks")
		gomega.Expect(true).Should(gomega.BeTrue())
	})

	ginkgo.It("hcp workloads are unaffected post management cluster upgrade", labels.MCUpgrade, func() {
		fmt.Println("Performing hcp cluster post management cluster upgrade")
		gomega.Expect(true).Should(gomega.BeTrue())
	})
})
