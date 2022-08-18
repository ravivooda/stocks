package utils

import (
	"fmt"
	"runtime"
	"time"
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func RetryFetching(original func() ([][]string, error), maxRetries int, sleep time.Duration) ([][]string, error) {
	var records [][]string
	var err error
	for maxRetries > 0 {
		records, err = original()
		if err == nil {
			break
		}
		fmt.Printf("retrying in %v\n", sleep)
		maxRetries -= 1
	}
	return records, err
}
