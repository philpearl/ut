// +build ignore
package main

import (
	"github.com/philpearl/ut"
)

type MockFred struct {
	ut.CallTracker
}

func (i *MockFred) Thing(a int, b ...string) error {
	params := []interface{}{}
	params[0] = a
	for i, p := range b {
		params[1+i] = p
	}
	r := i.TrackCall("Thing", params...)
	var r_0 error
	if r[0] != nil {
		r_0 = r[0].(error)
	}
	return r_0
}

type Fred interface {
	sanit(blah string)
	iit(fred interface{})
	many(things ...string)
	doit(blah string) int
	donit(blah, fah string) (int, error)
	adonit(blah, fah string, brian func(int) error) (an int, err error)
}
