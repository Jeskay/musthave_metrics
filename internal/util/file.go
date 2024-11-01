package util

import "os"

func IsValidPath(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	if err := os.WriteFile(path, make([]byte, 0), 0644); err == nil {
		os.Remove(path)
		return true
	}
	return false
}
