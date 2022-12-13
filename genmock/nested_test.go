package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNestedInterfaces(t *testing.T) {
	t.Run("exported", func(t *testing.T) {
		generateMock(&options{
			packagePath:   "testcode",
			ifName:        "Interface4",
			outfile:       filepath.Join("gentestfile", "exported.go"),
			targetPackage: "testcode",
			mockName:      "MockInterface4", // creates NewMockInterface4.
		})
		assertFileContent(t, "gentestfile/exported.go", "gentestfile/exported.golden")
	})
	t.Run("unexported", func(t *testing.T) {
		generateMock(&options{
			packagePath:   "testcode",
			ifName:        "Interface4",
			outfile:       filepath.Join("gentestfile", "unexported.go"),
			targetPackage: "testcode",
			mockName:      "mockInterface4", // creates newMockInterface4.
		})
		assertFileContent(t, "gentestfile/unexported.go", "gentestfile/unexported.golden")
	})
}

func assertFileContent(t *testing.T, actFile, expFile string) {
	act, err := os.ReadFile(actFile)
	if err != nil {
		t.Fatalf("Error opening file: %s", err)
	}

	exp, err := os.ReadFile(expFile)
	if err != nil {
		t.Fatalf("Error opening file: %s", err)
	}

	if diff := cmp.Diff(string(exp), string(act)); diff != "" {
		t.Fatalf("Expected contents of file %s to equal %s but they were not. %s", actFile, expFile, diff)
	}
}
