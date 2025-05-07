package gofumpt

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mirecl/golangci-lint/v2/pkg/config"
	"github.com/mirecl/golangci-lint/v2/pkg/goanalysis"
	"github.com/mirecl/golangci-lint/v2/pkg/goformatters"
	gofumptbase "github.com/mirecl/golangci-lint/v2/pkg/goformatters/gofumpt"
	"github.com/mirecl/golangci-lint/v2/pkg/golinters/internal"
)

const linterName = "gofumpt"

func New(settings *config.GoFumptSettings) *goanalysis.Linter {
	a := goformatters.NewAnalyzer(
		internal.LinterLogger.Child(linterName),
		"Checks if code and import statements are formatted, with additional rules.",
		gofumptbase.New(settings, settings.LangVersion),
	)

	return goanalysis.NewLinter(
		a.Name,
		a.Doc,
		[]*analysis.Analyzer{a},
		nil,
	).WithLoadMode(goanalysis.LoadModeSyntax)
}
