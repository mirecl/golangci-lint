package golines

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mirecl/golangci-lint/v2/pkg/config"
	"github.com/mirecl/golangci-lint/v2/pkg/goanalysis"
	"github.com/mirecl/golangci-lint/v2/pkg/goformatters"
	golinesbase "github.com/mirecl/golangci-lint/v2/pkg/goformatters/golines"
	"github.com/mirecl/golangci-lint/v2/pkg/golinters/internal"
)

const linterName = "golines"

func New(settings *config.GoLinesSettings) *goanalysis.Linter {
	a := goformatters.NewAnalyzer(
		internal.LinterLogger.Child(linterName),
		"Checks if code is formatted, and fixes long lines",
		golinesbase.New(settings),
	)

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeSyntax)
}
