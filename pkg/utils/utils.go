package utils

import (
	"errors"
	"fmt"
	"os"
)

// PathExists determines if a path exists
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return true, fmt.Errorf("warning: exists but another error happened (debug): %s", err)
	}
	return true, nil
}
