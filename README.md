# escape

## Building

	# Install the check and escape binaries used by generate.sh
	go install ./...

	# (Optional) move escape-go to a directory in your PATH.

## Running

	# From the root directory of a package that uses escape() in .ego files.

    escape FILE.ego > FILE.go

## Translation

Escape looks for lines that look like:

	check := escape(&err)

and expands these in-place so that the function returned by escape can be
used to trigger an early return from the enclosing function.

Only short variable declarations are currently expanded.

## Compilation

The escape analysis output of the go compiler can be used to catch
potential misuses of escape:

    go build -gcflags -m 2>&1 |
    grep 'func literal escapes to heap' |
    grep -Ff <(find . -name '*.go' |
        xargs -n1 grep -EHn '\W+escapePanicFlag := false$' |
        awk -F: '{ print $1":"$2+1 }') |
    sed -e 's/func literal escapes to heap/possible misuse of escape/g'

The line numbers reported refer to the generated code.
