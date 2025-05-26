package main

import (
	"github.com/MxTrap/metrics/osexitanalyzer"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/ultraware/whitespace"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/quickfix/qf1006"
	"honnef.co/go/tools/simple/s1005"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1008"
)

// main запускает staticlint.
// Для запуска необходимо указать путь до директории в которой необходимо провести проверку.
// Например, находясь в директрии со staticlint ввести в терминале go run main.go */metrics/..., звездочкой заменяется часть абсолютного пути
func main() {
	checks := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers)+15)
	checks = append(
		checks,
		// printf.Analyzer check consistency of Printf format strings and arguments
		printf.Analyzer,
		// appends.Analyzer that detects if there is only one variable in append.
		appends.Analyzer,
		// assign.Analyzer reports assignments of the form x = x or a[i] = a[i]. These are almost always useless, and even when they aren't they are usually a mistake.
		assign.Analyzer,
		// bools.Analyzer that detects common mistakes involving boolean operators
		bools.Analyzer,
		// defers.Analyzer that checks for common mistakes in defer statements
		defers.Analyzer,
		// loopclosure.Analyzer that checks for references to enclosing loop variables from within nested functions
		loopclosure.Analyzer,
		// lostcancel.Analyzer that checks for failure to call a context cancellation function
		lostcancel.Analyzer,
		// shadow.Analyzer that checks for shadowed variables
		shadow.Analyzer,
		// nilfunc.Analyzer that checks for useless comparisons against nil
		nilfunc.Analyzer,
		// structtag.Analyzer that checks struct field tags are well formed
		structtag.Analyzer,
		// unreachable.Analyzer that checks for unreachable code
		unreachable.Analyzer,
		// Drop unnecessary use of the blank identifier
		s1005.Analyzer,
		// A function’s error value should be its last return value
		st1008.Analyzer,
		// Lift if+break into loop condition
		qf1006.Analyzer,
		// errcheck is a program for checking for unchecked errors in Go code.
		errcheck.Analyzer,
		// Whitespace is a linter that checks for unnecessary newlines at the start and end of functions, if, for, etc.
		whitespace.NewAnalyzer(nil),
		// Анализатор, который проверяет прямой вызов os.Exit в функции main
		osexitanalyzer.Analyzer,
	)
	// Package staticcheck contains analyzes that find bugs and performance issues. Barring the rare false positive, any code flagged by these analyzes needs to be fixed.
	for _, a := range staticcheck.Analyzers {
		checks = append(checks, a.Analyzer)
	}

	multichecker.Main(
		checks...,
	)
}
