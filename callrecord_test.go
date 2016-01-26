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
	return r[0].(int), NilOrError(r[1])
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

func TestRecord(t *testing.T) {
	m := NewMockReader(t)

	m.RecordCall("Read", 1, nil)

	n, err := m.Read([]byte("cherry"))
	if n != 1 {
		t.Fatalf("n should be 1")
	}
	if err != nil {
		t.Fatalf("err should be nil")
	}
	n, err = m.Read([]byte("bomb"))
	if n != 1 {
		t.Fatalf("n should be 1")
	}
	if err != nil {
		t.Fatalf("err should be nil")
	}

	m.AssertDone()
	params, ok := m.GetRecordedParams("Read")
	if !ok {
		t.Fatalf("oops")
	}
	if len(params) != 2 {
		t.Fatalf("should have 2 calls")
	}

	if len(params[0]) != 1 {
		t.Fatalf("should have 1 param")
	}
	if string(params[0][0].([]byte)) != "cherry" {
		t.Fatal("grief!")
	}

	if string(params[1][0].([]byte)) != "bomb" {
		t.Fatal("grief!")
	}

}
