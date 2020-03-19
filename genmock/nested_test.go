package main

import (
	"path/filepath"
	"testing"
)

func TestNestedInterfaces(t *testing.T) {

	/*
		This test will generate a file in the gentestfile directory
	*/

	o := &options{
		packagePath:   "testcode",
		ifName:        "Interface4",
		outfile:       filepath.Join("gentestfile", "generated.go"),
		targetPackage: "testcode",
		mockName:      "mockInterface4",
	}
	generateMock(o)
}
