package errors

import (
	"errors"
)

var New = errors.New

func Check(f func(err error)) func(err error) {
	return func(err error) {
		if err != nil {
			f(err)
		}
	}
}

func Handle(err *error, f func()) {
	if *err != nil {
		f()
	}
}
