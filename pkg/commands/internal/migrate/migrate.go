package migrate

import (
	"github.com/mirecl/golangci-lint/v2/pkg/commands/internal/migrate/ptr"
	"github.com/mirecl/golangci-lint/v2/pkg/commands/internal/migrate/versionone"
	"github.com/mirecl/golangci-lint/v2/pkg/commands/internal/migrate/versiontwo"
)

func ToConfig(old *versionone.Config) *versiontwo.Config {
	return &versiontwo.Config{
		Version:    ptr.Pointer("2"),
		Linters:    toLinters(old),
		Formatters: toFormatters(old),
		Issues:     toIssues(old),
		Output:     toOutput(old),
		Severity:   toSeverity(old),
		Run:        toRun(old),
	}
}
