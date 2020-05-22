package helper

import "os"

// IsFileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
