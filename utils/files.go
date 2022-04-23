package utils

import (
	"errors"
	"os"
)

func MakeDirs(dirs []string) (bool, error) {
	for _, dirPathAddr := range dirs {
		if _, err := os.Stat(dirPathAddr); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dirPathAddr, os.ModePerm)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}
