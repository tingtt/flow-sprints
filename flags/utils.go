package flags

import (
	"os"
	"strconv"
)

// Get uint env variable
func getUintEnv(key string, fallback uint) uint {
	// Get env
	if value, ok := os.LookupEnv(key); ok {
		// parse to uint
		var intValue, err = strconv.ParseUint(value, 10, 16)
		if err == nil {
			return uint(intValue)
		}
	}
	// Use fallbacn when env using `key` does not exist or failed to parse
	return fallback
}

// Get string env variable
func getEnv(key, fallback string) string {
	// Get env
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	// Use fallbacn when env using `key` does not exist
	return fallback
}
