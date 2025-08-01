package gostrictenum

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestEnumLinter(t *testing.T) {
	// Get the absolute path to the testdata directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testdata := filepath.Join(wd, "testdata")

	// Create the required directory structure for analysistest
	srcDir := filepath.Join(testdata, "src", "enumlinter")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Move test files to the required location
	testFiles := []string{"test.go", "comprehensive_test.go"}
	for _, file := range testFiles {
		src := filepath.Join(testdata, file)
		dst := filepath.Join(srcDir, file)

		// Read the source file
		content, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}

		// Write to destination
		err = os.WriteFile(dst, content, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Use the analysistest package to test our analyzer
	linter := GoStrictEnumLinter{}
	analyzers, err := linter.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, testdata, analyzers[0])
}
