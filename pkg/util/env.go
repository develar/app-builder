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
