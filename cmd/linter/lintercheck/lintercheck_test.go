package lintercheck_test

import (
	"testing"

	"github.com/KV2013/url-shortner-go/cmd/linter/lintercheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), lintercheck.Analyzer, "./...")
}
