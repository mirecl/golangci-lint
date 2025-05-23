package nolintlint

import (
	"fmt"
	"sync"

	"golang.org/x/tools/go/analysis"

	"github.com/mirecl/golangci-lint/v2/pkg/config"
	"github.com/mirecl/golangci-lint/v2/pkg/goanalysis"
	"github.com/mirecl/golangci-lint/v2/pkg/golinters/internal"
	nolintlint "github.com/mirecl/golangci-lint/v2/pkg/golinters/nolintlint/internal"
	"github.com/mirecl/golangci-lint/v2/pkg/lint/linter"
)

const LinterName = nolintlint.LinterName

func New(settings *config.NoLintLintSettings) *goanalysis.Linter {
	var mu sync.Mutex
	var resIssues []goanalysis.Issue

	var needs nolintlint.Needs
	if settings.RequireExplanation {
		needs |= nolintlint.NeedsExplanation
	}
	if settings.RequireSpecific {
		needs |= nolintlint.NeedsSpecific
	}
	if !settings.AllowUnused {
		needs |= nolintlint.NeedsUnused
	}

	lnt, err := nolintlint.NewLinter(needs, settings.AllowNoExplanation)
	if err != nil {
		internal.LinterLogger.Fatalf("%s: create analyzer: %v", nolintlint.LinterName, err)
	}

	analyzer := &analysis.Analyzer{
		Name: nolintlint.LinterName,
		Doc:  goanalysis.TheOnlyanalyzerDoc,
		Run: func(pass *analysis.Pass) (any, error) {
			issues, err := lnt.Run(pass)
			if err != nil {
				return nil, fmt.Errorf("linter failed to run: %w", err)
			}

			if len(issues) == 0 {
				return nil, nil
			}

			mu.Lock()
			resIssues = append(resIssues, issues...)
			mu.Unlock()

			return nil, nil
		},
	}

	return goanalysis.NewLinter(
		nolintlint.LinterName,
		"Reports ill-formed or insufficient nolint directives",
		[]*analysis.Analyzer{analyzer},
		nil,
	).WithIssuesReporter(func(*linter.Context) []goanalysis.Issue {
		return resIssues
	}).WithLoadMode(goanalysis.LoadModeSyntax)
}
