package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func annotate(abort func(error)) func(error, string, ...interface{}) {
	return func(err error, format string, args ...interface{}) {
		if err != nil {
			pc, file, line, ok := runtime.Caller(1)

			s := ""
			if ok {
				_, file := filepath.Split(file)

				f := runtime.FuncForPC(pc)

				name := "unknown"
				if f != nil {
					p := strings.Split(f.Name(), ".")
					if len(p) > 0 {
						name = p[len(p)-1]
					}
				}

				s = fmt.Sprintf("%s:%d:%s: ", file, line, name)
			}

			abort(fmt.Errorf(s+format+": %w", append(args, err)...))
		}
	}
}

func additionalContext() (err error) {
	var escape_hatch_1 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_1 = func(v error) { // generated by escape: annotate_test.ego:40:27
			*p = v

			if !panicking {
				panicking = true
				panic(&panicking)
			}
		}

		defer func() {
			if panicking {
				panicking = false

				r := recover()
				if r != &panicking {
					// It would be better if we could unrecover.
					// Or if panicking with the return value from recover, unrecovered.
					// Or if there was a mechanism specifically for this that did not
					// use panic/recover but was weaker than panic in the same way that
					// panic is weaker than runtime.Goexit.
					panic(r)
				}
			}
		}()
	}
	check := annotate(escape_hatch_1)

	check(errors.New("reason"), "call failed")

	return
}

func TestAdditionalErrors(t *testing.T) {
	err := additionalContext()
	if err == nil {
		t.Errorf("no wrapped error")
	}

	got := err.Error()
	want := "annotate_test.go:73:additionalContext: call failed: reason"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}