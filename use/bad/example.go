package main

import (
	"errors"
	"fmt"
)

var errFailure = errors.New("failure")

func failure() error {
	return errFailure
}

func success() error {
	return nil
}

func emptyFailure() (string, error) {
	return "", errFailure
}

func stringSuccess() (string, error) {
	return "hello", nil
}

func childCallsCheck(check func(error)) {
	check(success())

	check(failure())
}

func furtherComplicatedChain(found func(string)) {
	found("response")
}

func someComplicatedChain(found func(string)) {
	furtherComplicatedChain(found)
}

func nothingFancy() (err error) {
	var escape_hatch_1 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_1 = func(v error) { // generated by escape: example.ego:41:18
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
	check := escape_hatch_1

	check(success())

	check(failure())

	return
}

func nothingFancy2() (err error) {
	var escape_hatch_2 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_2 = func(v error) { // generated by escape: example.ego:51:18
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
	check := escape_hatch_2

	s, err := stringSuccess()
	check(err)

	println(s)

	_, err = emptyFailure()
	check(err)

	return
}

func passToChild() (err error) {
	var escape_hatch_3 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_3 = func(v error) { // generated by escape: example.ego:65:18
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
	check := escape_hatch_3

	childCallsCheck(check)

	return
}

func cannotPassToParent() (err error) {
	check := func() func(error) {
		var escape_hatch_4 func(error)
		{
			p := &err

			panicking := false

			escape_hatch_4 = func(v error) { // generated by escape: example.ego:74:24
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
		childCheck := escape_hatch_4

		return childCheck
	}()

	check(success())

	return
}

func cannotPassToParent2() (err error) {
	check := func() func(error) {
		var escape_hatch_5 func(error)
		{
			p := &err

			panicking := false

			escape_hatch_5 = func(v error) { // generated by escape: example.ego:86:19
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
		check := escape_hatch_5
		return check
	}()

	check(success())

	return
}

func cannotGiveCheckToAnotherGoroutine() (err error) {
	var escape_hatch_6 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_6 = func(v error) { // generated by escape: example.ego:96:18
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
	check := escape_hatch_6

	go childCallsCheck(check)

	return
}

func notJustForErrors() (response string) {
	var escape_hatch_7 func(string)
	{
		p := &response

		panicking := false

		escape_hatch_7 = func(v string) { // generated by escape: example.ego:104:18
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
	found := escape_hatch_7

	someComplicatedChain(found)

	return "not found"
}

func additionalContext() (err error) {
	var escape_hatch_8 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_8 = func(v error) { // generated by escape: example.ego:112:18
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
	check := escape_hatch_8
	annotate := func(err error, format string, args ...interface{}) {
		if err == nil {
			return
		}

		check(fmt.Errorf(format+": %w", append(args, err)...))
	}

	annotate(failure(), "call to failure() failed")

	return
}

func annotate(check func(error)) func(error, string, ...interface{}) {
	return func(err error, format string, args ...interface{}) {
		if err != nil {
			check(fmt.Errorf(format+": %w", append(args, err)...))
		}
	}
}

func additionalContext2() (err error) {
	var escape_hatch_9 func(error)
	{
		p := &err

		panicking := false

		escape_hatch_9 = func(v error) { // generated by escape: example.ego:135:27
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
	check := annotate(escape_hatch_9)

	check(failure(), "call to failure() failed")

	return
}

// Some comments that should stay put.
func main() {
	cannotGiveCheckToAnotherGoroutine()
	passToChild()
	cannotPassToParent2()
	notJustForErrors()
}