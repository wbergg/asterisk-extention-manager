package fax

import (
	"os"
	"path/filepath"
)

func CheckCallFileExists(spoolPath, callFileName string) bool {
	_, err := os.Stat(filepath.Join(spoolPath, callFileName))
	return err == nil
}
