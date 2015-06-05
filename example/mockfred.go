package example

import (
	"github.com/philpearl/ut"
	"testing"
)

type MockFred struct {
	ut.CallTracker
}

func NewMockFred(t *testing.T) *MockFred {
	return &MockFred{ut.NewCallRecords(t)}
}

func (m *MockFred) AddCall(name string, params ...interface{}) ut.CallTracker {
	m.CallTracker.AddCall(name, params...)
	return m
}

func (m *MockFred) SetReturns(params ...interface{}) ut.CallTracker {
	m.CallTracker.SetReturns(params...)
	return m
}
func (i *MockFred) sanit(blah string)	{ i.TrackCall("sanit", blah); return }

func (i *MockFred) iit(fred interface{}) {
	i.TrackCall("iit", fred)
	return
}

func (i *MockFred) many(things ...string) {
	params := make([]interface{}, 0+
		len(things,
		))
	for j, p := range things {

		params[0+j] = p
	}
	i.TrackCall("many", params...)
	return
}

func (i *MockFred) doit(blah string) int {
	r := i.TrackCall("doit", blah)
	var r_0 int
	if r[0] != nil {
		r_0 = r[0].(int)
	}
	return r_0
}
func (i *MockFred) donit(blah, fah string) (int, error) {
	r := i.TrackCall("donit", blah,

		fah)
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
func (i *MockFred) adonit(blah, fah string, brian func(int) error) (int, error) {
	r := i.TrackCall("adonit", blah,

		fah, brian,
	)
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
