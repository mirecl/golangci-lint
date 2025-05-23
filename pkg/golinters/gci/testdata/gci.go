//golangcitest:config_path testdata/gci.yml
package testdata

import (
	"golang.org/x/tools/go/analysis" // want "File is not properly formatted"
	"github.com/mirecl/golangci-lint/v2/pkg/config"
	"fmt"
	"errors"
	gcicfg "github.com/daixiang0/gci/pkg/config"
)

func GoimportsLocalTest() {
	fmt.Print(errors.New("x"))
	_ = config.Config{}
	_ = analysis.Analyzer{}
	_ = gcicfg.BoolConfig{}
}
