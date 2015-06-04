# UT

UT allows you to automatically generate mock implementations of go interfaces for easy and awesome unit testing.

[![GoDoc](https://godoc.org/github.com/philpearl/ut?status.svg)](https://godoc.org/github.com/philpearl/ut) 
[![Build Status](https://travis-ci.org/philpearl/ut.svg)](https://travis-ci.org/philpearl/ut)

## What's included
UT includes the following.

- Code to help you build mock implementations of interfaces. You can say what calls you expect on the mock objects, what the 
parameters should be and what each call should return.
- A tool (genmock) to automatically generate mocks from interface definitions.

The basic code is simple to understand and uses no magic. The auto code generation is not so simple to understand, but hopefully
should work without you needing to look at it!

The genmock tool currently only works if you can point it at the source file containing the interface type definition. I don't think it 
will be hard to get it to generate code for an interface in any package, but it isn't there yet.

## genmock

genmock's parameters are as follows

- filename: name of the file containing the interface definition. Must be specified.
- interface: name of the interface to create a mock for. Must be specified.
- mock: name of the mock object to create. Defaults to Mock<interface>.
- outfile: name of the file hold the mock definition. Defaults to mock<interface>.go in the current directory.
- package: name of the package to use in the mock definition. Must be specified.

Install genmock with `go install github.com/philpearl/ut/genmock`

You can then use it with go generate as follows. Add a go:generate comment as shown below (with no spaces within //go:generate), then run `go generate` to generate the files.

```go
package mypackage

type MyInterface {
	func MakeACall(param string) error
}

//go:generate genmock -filename=thisfile.go -interface=MyInterface -package mypackage
```

## Example

This example is implemented as a test in this package. It creates a mock io.Reader, and tests the function UnderTest(). In this case I've built the mock by
hand so you can see what kind of code genmock will generate.

```go
package ut

import (
	"io"
	"testing"
)

// For this test we implement a mock of the io.Reader interface
type MockReader struct {
	CallTracker
}

// NewMockReader is a convenience method for creating our mock
func NewMockReader(t *testing.T) *MockReader {
	return &MockReader{NewCallRecords(t)}
}

// Here we implement the Read method of our mock io.Reader. This
// records the parameters passed to the call and returns values
// specified by the test
func (m *MockReader) Read(p []byte) (n int, err error) {
	r := m.TrackCall("Read", p)
	var r_0 int
	if r[0] != nil {
	    r_0 = r[0].(int)
	}
	var r_1 error
	if r[1] != nil {
		r_1 = r[1].(error)
	}
	return r_0, r_1
}

// This is the function we're going to test.
func UnderTest(r io.Reader) bool {
	p := make([]byte, 10)
	n, _ := r.Read(p)

	return n >= 1 && p[0] == 37
}

func TestUnderTest(t *testing.T) {

	// Define the tests we're going to run.
	tests := []struct {
		bytezero byte
		n        int
		expRet   bool
	}{
		{bytezero: 37, n: 1, expRet: true},
		{bytezero: 37, n: 2, expRet: true},
		{bytezero: 38, n: 2, expRet: false},
		{bytezero: 0, n: 2, expRet: false},
		{bytezero: 37, n: 0, expRet: false},
		{bytezero: 0, n: 0, expRet: false},
	}

	for _, test := range tests {
		// Set up the mock
		m := NewMockReader(t)

		// Parameters for AddCall can either be: values, which are compared against the actual parameter;
		// or functions, which can check and act on the parameter as they like
		checkReadParam := func(p interface{}) {
			buf := p.([]byte)
			if len(buf) != 10 {
				t.Fatalf("should have read 10 bytes")
			}
			buf[0] = test.bytezero
		}

		// Note the calls we expect to happen when we run our test
		m.AddCall("Read", checkReadParam).SetReturns(test.n, error(nil))

		// Test the function
		if UnderTest(m) != test.expRet {
			t.Fatalf("return not as expected")
		}

		// Check the method calls we expected actually happened
		m.AssertDone()
	}
}

```