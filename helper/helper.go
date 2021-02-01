package helper

import (
	"fmt"
	"os"
)

// IsFileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// PanicIfErrf panic and reports error message with args
func PanicIfErrf(err error, msg string, args ...interface{}) {
	if err != nil {
		fmt.Printf(msg, args...)
		fmt.Println()
		panic(err)
	}
}

// PanicIfErr panic and reports error message
func PanicIfErr(err error, msg string) {
	if err != nil {
		fmt.Printf(msg)
		fmt.Println()
		panic(err)
	}
}
