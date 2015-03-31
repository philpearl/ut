package ut

import (
	"reflect"
	"testing"
)

type CallTracker interface {
	// AddCall() is used by tests to add an expected call to the tracker
	AddCall(name string, params ...interface{}) CallTracker

	// SetReturns() is called immediately after AddCall() to set the return values for the call.
	SetReturns(returns ...interface{}) CallTracker

	// TrackCall() is called within mocks to track a call to the Mock. It returns the return values
	// registered via SetReturns()
	TrackCall(name string, params ...interface{}) []interface{}

	// AssertDone() should be called at the end of a test to confirm all the expected calls have
	// been made
	AssertDone()
}

type callRecord struct {
	name    string
	params  []interface{}
	returns []interface{}
}

func (e *callRecord) assert(t *testing.T, name string, params ...interface{}) {
	if name != e.name {
		t.Fatalf("Expected call to  %s, got call to %s", e.name, name)
	}
	if len(params) != len(e.params) {
		t.Fatalf("Call to (%s) expected %d params, got %d", name, len(e.params), len(params))
	}
	for i, ap := range params {
		ep := e.params[i]
		switch ep := ep.(type) {
		case func(actual interface{}):
			ep(ap)
		default:
			if !reflect.DeepEqual(ap, ep) {
				t.Fatalf("Call to (%s) parameter %d got %v (%T) does not match expected %v (%T)", name, i, ap, ap, ep, ep)
			}
		}
	}
}

type callRecords struct {
	t       *testing.T
	calls   []callRecord
	current int
}

func NewCallRecords(t *testing.T) CallTracker {
	return &callRecords{
		t: t,
	}
}

func (cr *callRecords) AddCall(name string, params ...interface{}) CallTracker {
	cr.calls = append(cr.calls, callRecord{name: name, params: params})
	return cr
}

func (cr *callRecords) SetReturns(returns ...interface{}) CallTracker {
	cr.calls[len(cr.calls)-1].returns = returns
	return cr
}

func (cr *callRecords) TrackCall(name string, params ...interface{}) []interface{} {
	expectedCall := cr.calls[cr.current]
	expectedCall.assert(cr.t, name, params...)
	cr.current += 1
	return expectedCall.returns
}

func (cr *callRecords) AssertDone() {
	if cr.current < len(cr.calls) {
		cr.t.Fatalf("Only %d of %d expected calls made", cr.current, len(cr.calls))
	}
}
