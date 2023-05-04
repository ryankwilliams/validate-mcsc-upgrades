package validate_mcsc_upgrades_test

import (
	"testing"

	_ "github.com/ryankwilliams/validate-mcsc-upgrades/tests"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestValidateMcscUpgrades(t *testing.T) {
	RegisterFailHandler(Fail)

	suiteConfig, reporterConfig := GinkgoConfiguration()
	suiteConfig.LabelFilter = "ApplyHCPWorkloads || RemoveHCPWorkloads || MCUpgrade || SCUpgrade"
	reporterConfig.JUnitReport = "junit.xml"

	RunSpecs(t, "ValidateMcScUpgrades Suite", suiteConfig, reporterConfig)
}
