package utils

import (
	"github.com/pkg/errors"
)

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicErrWithExtraMessage(err error, message string) {
	if err != nil {
		panic(errors.Wrap(err, message))
	}
}