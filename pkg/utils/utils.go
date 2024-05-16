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

// Copy returns a copy of a list slice
// This is intended for evaluating contender clusters
// where we change the list (and need to reuse it later)
func Copy(list []string) []string {
	copied := make([]string, len(list))
	copy(copied, list)
	return copied
}

// Diff returns the difference between two sets of strings
func Diff(one, two []string) []string {
	var difference []string
	lookup := make(map[string]struct{}, len(two))
	for _, item := range two {
		lookup[item] = struct{}{}
	}
	for _, item := range one {
		if _, found := lookup[item]; !found {
			difference = append(difference, item)
		}
	}
	return difference
}
