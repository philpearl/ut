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

	// RecordCall() is called to indicate calls to the named mock method should
	// be recorded rather than asserted.  The parameters to any call to the
	// named method will be recorded and may be retrieved via GetRecordedParams.
	// The returns from the method are also specified on this call and must be
	// the same each time.
	// Note that the ordering of recorded calls relative to other calls is not
	// tracked.
	RecordCall(name string, returns ...interface{}) CallTracker

	// GetRecordedParams returns the sets of parameters passed to a call captured
	// via RecordCall
	GetRecordedParams(name string) ([][]interface{}, bool)
}

type callRecord struct {
	name        string
	params      []interface{}
	returns     []interface{}
	addLocation codeLocation
}

type codeLocation struct {
	file string
	line int
}

func (e *callRecord) assert(t testing.TB, name string, params ...interface{}) {
	if name != e.name {
		t.Logf("Expected call to %s%s%s", e.name, paramsToString(e.params), codeLocationToString(e.addLocation))
		t.Logf(" got call to %s%s", name, paramsToString(params))
		showStack(t)
		t.Fail()
		return
	}
	if len(params) != len(e.params) {
		t.Logf("Call to (%s) unexpected parameters%s", name, codeLocationToString(e.addLocation))
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
				t.Logf("Call to %s parameter %d unexpected%s", name, i, codeLocationToString(e.addLocation))
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

func codeLocationToString(c codeLocation) string {
	return fmt.Sprintf(" (Added in %s:%d)", c.file, c.line)
}

// recording tracks calls actually made to the mock. It is used only when the
// user choses to record calls for a method rather than assert them
type recording struct {
	// The returned values are the same for each call to a recorded method.
	returns []interface{}
	// We record the parameters from each call to the method.
	params [][]interface{}
}

type callRecords struct {
	sync.Mutex
	t       testing.TB
	calls   []callRecord
	records map[string]*recording
	current int
}

// NewCallRecords creates a new call tracker
func NewCallRecords(t testing.TB) CallTracker {
	return &callRecords{
		t:       t,
		records: make(map[string]*recording),
	}
}

func (cr *callRecords) AddCall(name string, params ...interface{}) CallTracker {
	c := callRecord{name: name, params: params}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		c.addLocation = codeLocation{
			file: file,
			line: line,
		}
	}
	cr.calls = append(cr.calls, c)
	return cr
}

func (cr *callRecords) RecordCall(name string, returns ...interface{}) CallTracker {
	cr.records[name] = &recording{
		returns: returns,
		params:  make([][]interface{}, 0),
	}
	return cr
}

func (cr *callRecords) SetReturns(returns ...interface{}) CallTracker {
	cr.calls[len(cr.calls)-1].returns = returns
	return cr
}

func (cr *callRecords) TrackCall(name string, params ...interface{}) []interface{} {
	cr.Lock()
	defer cr.Unlock()
	if record, ok := cr.records[name]; ok {
		// Call is to be recorded, not asserted
		record.params = append(record.params, params)
		return record.returns
	}
	// Call is to be asserted
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
		// We don't call Fatalf or FailNow because that may mask other errors if this AssertDone
		// is called from a defer
		missed := &bytes.Buffer{}
		for i, call := range cr.calls[cr.current:] {
			if i != 0 {
				missed.WriteString(", ")
			}
			missed.WriteString(call.name)
		}

		cr.t.Errorf("Only %d of %d expected calls made. Missed calls to %s", cr.current, len(cr.calls), missed)
	}
}

func (cr *callRecords) GetRecordedParams(name string) ([][]interface{}, bool) {
	cr.Lock()
	defer cr.Unlock()
	record, ok := cr.records[name]
	if ok {
		return record.params, true
	}
	return nil, false
}

// NilOrError is a utility function for returning err from mocked methods
func NilOrError(val interface{}) error {
	if val == nil {
		return nil
	}
	return val.(error)
}
