package utils

import (
	"fmt"
	"github.com/pkg/errors"
	"runtime/debug"
)

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func LogErr(err error) {
	debug.PrintStack()
	fmt.Printf("%+v\n", err)
}

func PanicErrWithExtraMessage(err error, message string) {
	if err != nil {
		panic(errors.Wrap(err, message))
	}
}
