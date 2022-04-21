package utils

import (
	"errors"
	"os"
)

func MakeDir(dirPathAddr string) (bool, error) {
	if _, err := os.Stat(dirPathAddr); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(dirPathAddr, os.ModePerm)
		if err != nil {
			return false, err
		}
	}
	return false, nil
}
