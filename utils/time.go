package utils

import "time"

func TodayDate() string {
	return time.Now().Format("01-02-2006")
}
