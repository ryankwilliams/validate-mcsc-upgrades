package provider

func IdentifyClusterVersion() (string, error) {
	return "install version", nil
}

func DetermineUpgradeVersion(currentVersion string) (*string, error) {
	version := "upgrade version"
	return &version, nil
}
