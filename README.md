# Adding Escape Hatches to Go

## Installing

	# Install the check and escape binaries used by generate.sh.
	go install ./...

	# (Optional) move expand-escape to a directory in your PATH.

## Running

	# From the root directory of a package that uses escape() in .ego files.
    expand-escape && go build

### Translation

Escape looks for lines that look like,

	abort := escape(&err)

or,

	check := f(escape(&err))

in .ego files, and rewrites them (saving the result to a .go file) so that
the function returned by escape can be used to trigger an early return from
the function that called escape.

Only short variable declarations and escape() used as an argument are
currently expanded.

### Restrictions

The function returned by escape can only be called by a function with a call
chain rooted at the function that called escape. When other uses are detected
they are reported as errors.
