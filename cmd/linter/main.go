package main

import (
	"github.com/KV2013/url-shortner-go/cmd/linter/lintercheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(lintercheck.Analyzer)
}
