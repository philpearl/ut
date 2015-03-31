package ut

import (
	"reflect"
	"testing"
)

type CallRecord struct {
	Name   string
	Params []interface{}
}

type CallRecords struct {
	Calls []CallRecord
}

func (cr *CallRecords) AddCall(name string, params ...interface{}) *CallRecords {
	cr.Calls = append(cr.Calls, CallRecord{Name: name, Params: params})
	return cr
}

func (cr *CallRecords) AssertEquals(t *testing.T, exp *CallRecords) {
	if len(cr.Calls) != len(exp.Calls) {
		t.Fatalf("Number of calls %d does nt match expected %d", len(cr.Calls), len(exp.Calls))
	}

	for i, c := range cr.Calls {
		e := exp.Calls[i]
		if c.Name != e.Name {
			t.Fatalf("Call %d expected %s, got %s", i, e.Name, c.Name)
		}
		if len(c.Params) != len(e.Params) {
			t.Fatalf("Call %d (%s) expected %d params, got %d", i, c.Name, len(e.Params), len(c.Params))
		}
		for j, ap := range c.Params {
			ep := e.Params[j]
			switch ep := ep.(type) {
			case func(actual interface{}):
				ep(ap)
			default:
				if !reflect.DeepEqual(ap, ep) {
					t.Fatalf("Call %d (%s) parameter %d got %v (%T) does not match expected %v (%T)", i, c.Name, j, ap, ap, ep, ep)
				}
			}
		}
	}
}
