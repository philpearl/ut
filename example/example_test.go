package example

import (
	"testing"
)

func TestDoSomething(t *testing.T) {
	mf := NewMockFred(t)

	mf.AddCall("sanit", "cheese")
	mf.AddCall("doit", "lemons").SetReturns(5)
	mf.AddCall("many", "a", "b")

	DoSomething(mf)

	// Check that all the calls are made
	mf.AssertDone()
}

func TestDoSomethingElse(t *testing.T) {
	mf := NewMockFred(t)

	mf.AddCall("sanit", "cheese")
	// If doit returns a value less than 5 DoSomething doesn't call many
	mf.AddCall("doit", "lemons").SetReturns(3)

	DoSomething(mf)

	// Check that all the calls are made
	mf.AssertDone()
}
