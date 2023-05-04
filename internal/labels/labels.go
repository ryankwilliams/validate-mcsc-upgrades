package labels

import "github.com/onsi/ginkgo/v2"

var (
	ApplyHCPWorkloads     = ginkgo.Label("ApplyHCPWorkloads")
	MCUpgrade             = ginkgo.Label("MCUpgrade")
	MCUpgradeHealthChecks = ginkgo.Label("MCUpgradeHealthChecks")
	RemoveHCPWorkloads    = ginkgo.Label("RemoveHCPWorkloads")
	SCUpgrade             = ginkgo.Label("SCUpgrade")
	SCUpgradeHealthChecks = ginkgo.Label("SCUpgradeHealthChecks")
)
