package labels

import "github.com/onsi/ginkgo/v2"

var (
	MCUpgrade = ginkgo.Label(
		"ManagementClusterUpgrade",
		"MCUpgrade",
	)

	SCUpgrade = ginkgo.Label(
		"ServiceClusterUpgrade",
		"SCUpgrade",
	)

	MCUpgradeHealthChecks = ginkgo.Label(
		"ManagementClusterUpgradeHealthChecks",
		"MCUpgradeHealthChecks",
	)

	SCUpgradeHealthChecks = ginkgo.Label(
		"ServiceClusterUpgradeHealthChecks",
		"SCUpgradeHealthChecks",
	)

	ApplyHCPWorkloads = ginkgo.Label("ApplyHCPWorkloads")

	RemoveHCPWorkloads = ginkgo.Label("RemoveHCPWorkloads")
)
