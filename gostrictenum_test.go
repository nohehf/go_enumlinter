package gostrictenum

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestEnumLinter(t *testing.T) {
	testdata := analysistest.TestData()

	// Use the analysistest package to test our analyzer
	linter := GoStrictEnumLinter{}
	analyzers, err := linter.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, testdata, analyzers[0], "testdata/src/gostrictenum")
}
