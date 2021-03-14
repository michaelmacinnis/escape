// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted from:
// callgraph: a tool for reporting the call graph of a Go program.
// See: https://pkg.go.dev/golang.org/x/tools/cmd/callgraph
package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/cha"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/pointer"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// flags
var (
	algoFlag = flag.String("algo", "rta",
		`Call graph construction algorithm (static, cha, rta, pta)`)

	testFlag = flag.Bool("test", false,
		"Loads test code (*_test.go) for imported packages")
)

func init() {
	flag.Var((*buildutil.TagsFlag)(&build.Default.BuildTags), "tags", buildutil.TagsFlagDoc)
}

const Usage = `callgraph: display the call graph of a Go program.

Usage:

  callgraph [-algo=static|cha|rta|pta] [-test] [-format=...] package...

Flags:

-algo      Specifies the call-graph construction algorithm, one of:

            static      static calls only (unsound)
            cha         Class Hierarchy Analysis
            rta         Rapid Type Analysis
            pta         inclusion-based Points-To Analysis

           The algorithms are ordered by increasing precision in their
           treatment of dynamic calls (and thus also computational cost).
           RTA and PTA require a whole program (main or test), and
           include only functions reachable from main.

-test      Include the package's tests in the analysis.
`

func main() {
	flag.Parse()
	if err := doCallgraph("", "", *algoFlag, *testFlag, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "callgraph: %s\n", err)
		os.Exit(1)
	}
}

func doCallgraph(dir, gopath, algo string, tests bool, args []string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, Usage)
		return nil
	}

	type hatch struct {
		base string
		line int
	}

	watches := map[hatch]struct{}{}

	for {
		base := ""
		line := 0

		n, err := fmt.Scanf("%s %d\n", &base, &line)
		if err != nil || n != 2 {
			break
		}

		watches[hatch{base, line}] = struct{}{}
	}

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: tests,
		Dir:   dir,
	}
	if gopath != "" {
		cfg.Env = append(os.Environ(), "GOPATH="+gopath) // to enable testing
	}
	initial, err := packages.Load(cfg, args...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	prog, pkgs := ssautil.AllPackages(initial, 0)
	prog.Build()

	// -- call graph construction ------------------------------------------

	var cg *callgraph.Graph

	switch algo {
	case "static":
		cg = static.CallGraph(prog)

	case "cha":
		cg = cha.CallGraph(prog)

	case "pta":
		// Set up points-to analysis log file.
		mains, err := mainPackages(pkgs)
		if err != nil {
			return err
		}
		config := &pointer.Config{
			Mains:          mains,
			BuildCallGraph: true,
		}
		ptares, err := pointer.Analyze(config)
		if err != nil {
			return err // internal error in pointer analysis
		}
		cg = ptares.CallGraph

	case "rta":
		mains, err := mainPackages(pkgs)
		if err != nil {
			return err
		}
		var roots []*ssa.Function
		for _, main := range mains {
			roots = append(roots, main.Func("init"), main.Func("main"))
		}
		rtares := rta.Analyze(roots, true)
		cg = rtares.CallGraph

		// NB: RTA gives us Reachable and RuntimeTypes too.

	default:
		return fmt.Errorf("unknown algorithm: %s", algo)
	}

	cg.DeleteSyntheticNodes()

	// -- output------------------------------------------------------------

	callers := map[string]map[string]struct{}{}
	hatches := map[string]string{}

	if err := callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		data := Edge{fset: prog.Fset}

		data.edge = edge
		data.Caller = edge.Caller.Func
		data.CallerPosition = prog.Fset.Position(data.Caller.Pos())
		data.Callee = edge.Callee.Func
		data.CalleePosition = prog.Fset.Position(data.Callee.Pos())

		parent := data.Caller.Parent()
		if parent != nil {
			data.CallerParentPosition = prog.Fset.Position(parent.Pos())
		}

		parent = data.Callee.Parent()
		if parent != nil {
			data.CalleeParentPosition = prog.Fset.Position(parent.Pos())
		}

		if _, ok := edge.Site.(*ssa.Go); !ok {
		from, ok := callers[data.CalleePosition.String()]
		if !ok {
			from = map[string]struct{}{}
			callers[data.CalleePosition.String()] = from
		}
		from[data.CallerPosition.String()] = struct{}{}

		var h *hatch

		pos := data.CalleePosition
		for k := range watches {
			if strings.HasSuffix(pos.Filename, k.base) && pos.Line == k.line {
				h = &k
				break
			}
		}

		if h != nil {
			hatches[pos.String()] = data.CalleeParentPosition.String()
			delete(watches, *h)
		}
	}

		return nil
	}); err != nil {
		return err
	}

	for hatch, parent := range hatches {
		if from, ok := callers[hatch]; ok {
			if !valid(callers, parent, from) {
				fmt.Printf("%s\n", hatch)
			}
		}
	}

	return nil
}

func valid(callers map[string]map[string]struct{}, parent string, from map[string]struct{}) bool {
	for caller := range from {
		if caller != parent {
			if caller == "-" {
				return false
			}
			from, ok := callers[caller]
			if !ok || !valid(callers, parent, from) {
				return false
			}
		}
	}
	return true
}

// mainPackages returns the main packages to analyze.
// Each resulting package is named "main" and has a main function.
func mainPackages(pkgs []*ssa.Package) ([]*ssa.Package, error) {
	var mains []*ssa.Package
	for _, p := range pkgs {
		if p != nil && p.Pkg.Name() == "main" && p.Func("main") != nil {
			mains = append(mains, p)
		}
	}
	if len(mains) == 0 {
		return nil, fmt.Errorf("no main packages")
	}
	return mains, nil
}

type Edge struct {
	Caller *ssa.Function
	CallerPosition token.Position

	CallerParentPosition token.Position

	Callee *ssa.Function
	CalleePosition token.Position

	CalleeParentPosition token.Position

	edge     *callgraph.Edge
	fset     *token.FileSet
}

func (e *Edge) Dynamic() string {
	if e.edge.Site != nil {
		if _, ok := e.edge.Site.(*ssa.Go); ok {
			return "go"
		} else if _, ok = e.edge.Site.(*ssa.Defer); ok {
			return "defer"
		} else if e.edge.Site.Common().StaticCallee() == nil {
			return "dynamic"
		}
	}
	return "static"
}
