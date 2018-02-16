package util

import "os"

func GetEnvOrDefault(envName string, defaultValue string) string {
	result := os.Getenv(envName)
	if result == "" {
		return defaultValue
	} else {
		return result
	}
}

func IsEnvTrue(envName string) bool {
	value, ok := os.LookupEnv(envName)
	return ok && (value == "true" || value == "" || value == "1")
}
