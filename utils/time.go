package utils

import (
	"fmt"
	"time"
)

func TodayDate() string {
	return time.Now().Format("01-02-2006")
}

func Elapsed(what string) func() {
	start := time.Now()
	fmt.Printf("%s began!\n", what)
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}
