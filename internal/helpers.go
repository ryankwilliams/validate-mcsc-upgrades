package internal

import "os"

// getEnvVar gets environment variable value and returns default if unset
func GetEnvVar(key, value string) string {
	result, exist := os.LookupEnv(key)
	if exist {
		return result
	}
	return value
}
