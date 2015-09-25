// Package ut implements some testing utilities. So far it includes CallTracker, which helps you build
// mock implementations of interfaces.
package ut

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

// CallTracker is an interface to help build mocks.
//
// Build the CallTracker interface into your mocks. Use TrackCall within mock methods to track calls to the method
// and the parameters used.
// Within tests use AddCall to add expected method calls, and SetReturns to indicate what the calls will return.
//
// The tests for this package contain a full example.
//
//   type MyMock struct {ut.CallTracker}
//
//   func (m *MyMock) AFunction(p int) error {
//  	r := m.TrackCall("AFunction", p)
//      return NilOrError(r[0])
//   }
//
//   func Something(m Functioner) {
//      m.AFunction(37)
//   }
//
//   func TestSomething(t *testing.T) {
//  	m := &MyMock{NewCallRecords(t)}
//      m.AddCall("AFunction", 37).SetReturns(nil)
//
//      Something(m)
//
//      m.AssertDone()
//   }
type CallTracker interface {
	// AddCall() is used by tests to add an expected call to the tracker
	AddCall(name string, params ...interface{}) CallTracker

	// SetReturns() is called immediately after AddCall() to set the return
	// values for the call.
	SetReturns(returns ...interface{}) CallTracker

	// TrackCall() is called within mocks to track a call to the Mock. It
	// returns the return values registered via SetReturns()
	TrackCall(name string, params ...interface{}) []interface{}

	// AssertDone() should be called at the end of a test to confirm all
	// the expected calls have been made
	AssertDone()
}

type callRecord struct {
	name    string
	params  []interface{}
	returns []interface{}
}

func (e *callRecord) assert(t testing.TB, name string, params ...interface{}) {
	if name != e.name {
		t.Logf("Expected call to %s%s", e.name, paramsToString(e.params))
		t.Logf(" got call to %s%s", name, paramsToString(params))
		showStack(t)
		t.Fail()
		return
	}
	if len(params) != len(e.params) {
		t.Logf("Call to (%s) unexpected parameters", name)
		t.Logf(" expected %s", paramsToString(e.params))
		t.Logf("      got %s", paramsToString(params))
		showStack(t)
		t.FailNow()
		return
	}
	for i, ap := range params {
		ep := e.params[i]

		if ap == nil && ep == nil {
			continue
		}

		switch ep := ep.(type) {
		case func(actual interface{}):
			ep(ap)
		default:
			if !reflect.DeepEqual(ap, ep) {
				t.Logf("Call to %s parameter %d unexpected", name, i)
				t.Logf("  expected %#v (%T)", ep, ep)
				t.Logf("       got %#v (%T)", ap, ap)
				showStack(t)
				t.Fail()
			}
		}
	}
}

func showStack(t testing.TB) {
	pc := make([]uintptr, 10)
	n := runtime.Callers(4, pc)
	for i := 0; i < n; i++ {
		f := runtime.FuncForPC(pc[i])
		file, line := f.FileLine(pc[i])
		t.Logf("  %s (%s line %d", f.Name(), file, line)
	}
}

func paramsToString(params []interface{}) string {
	w := &bytes.Buffer{}
	w.WriteString("(")
	l := len(params)
	for i, p := range params {
		fmt.Fprintf(w, "%#v", p)
		if i < l-1 {
			w.WriteString(", ")
		}
	}
	w.WriteString(")")
	return w.String()
}

type callRecords struct {
	sync.Mutex
	t       testing.TB
	calls   []callRecord
	current int
}

// NewCallRecords creates a new call tracker
func NewCallRecords(t testing.TB) CallTracker {
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
	cr.Lock()
	defer cr.Unlock()
	if cr.current >= len(cr.calls) {
		cr.t.Logf("Unexpected call to %s%s", name, paramsToString(params))
		showStack(cr.t)
		cr.t.FailNow()
	}
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

// NilOrError is a utility function for returning err from mocked methods
func NilOrError(val interface{}) error {
	if val == nil {
		return nil
	}
	return val.(error)
}
