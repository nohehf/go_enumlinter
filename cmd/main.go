package main

import (
	"log"

	"github.com/0ri0nexe/gostrictenum"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	linter := gostrictenum.GoStrictEnumLinter{}
	analyzers, err := linter.BuildAnalyzers()
	if err != nil {
		log.Fatal(err)
	}
	singlechecker.Main(analyzers[0])
}
