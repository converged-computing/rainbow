package utils

import (
	"errors"
	"fmt"
	"math/rand"
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

// ShuffleJobs randomly shuffles jobs
func ShuffleJobs(jobs []int32) []int32 {
	for i := range jobs {
		j := rand.Intn(i + 1)
		jobs[i], jobs[j] = jobs[j], jobs[i]
	}
	return jobs
}
