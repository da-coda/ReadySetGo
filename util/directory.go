package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func IsWritable(path string) (bool, error) {
	testWritablePath := filepath.Clean(fmt.Sprintf("%s/writable", path))
	err := os.WriteFile(testWritablePath, []byte("I am writable"), os.ModeTemporary)
	if err != nil {
		return false, fmt.Errorf("unable to write to bin directory: %w", err)
	}
	err = os.Remove(testWritablePath)
	if err != nil {
		return false, fmt.Errorf("unable to delete files in bin directory: %w", err)
	}
	return true, nil
}
