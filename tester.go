package testutils

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"
)

// Tester wraps a *testing.T and an Index for iterative tests
type tester struct {
	t        *testing.T
	index    int
	indexSet bool
}

// Tester describes a testing context which can be modified to output an index for iterative testing
type Tester interface {
	// At sets index and returns this for convenient chaining as tt.At(i).CheckXXX()
	At(index int) Tester
	CheckEqual(expected interface{}, got interface{})
	CheckNumericGreater(expected interface{}, got interface{})
	CheckNumericLess(expected interface{}, got interface{})
	CheckEqualAndNoError(expected interface{}, got interface{}, gotError error)
	CheckNil(got interface{})
	CheckNotNil(got interface{})
	CheckError(got error)
	CheckNotError(got error)
	CheckTrue(got bool)
	CheckFalse(got bool)
	CheckAfter(expected, got time.Time, add ...time.Duration)
	CheckAfterOrEqual(expected, got time.Time, add ...time.Duration)
	CheckBefore(expected, got time.Time, add ...time.Duration)
	CheckBeforeOrEqual(expected, got time.Time, add ...time.Duration)
	CheckMatches(expected interface{}, got string)
	Fatalf(fmt string, args ...interface{})
	CheckTruef(predicate bool, fmt string, args ...interface{})
}

// NewTester returns a new tester that supports setting the Index
func NewTester(t *testing.T) Tester {
	return &tester{t: t}
}

func (tt *tester) At(index int) Tester {
	tt.indexSet = true
	tt.index = index
	return tt
}

func (tt *tester) unequalValues(e, g interface{}) {
	tt.t.Helper()
	tt.Fatalf("Expected: %T %v, got %T %v", e, e, g, g)
}

func (tt *tester) Fatalf(str string, args ...interface{}) {
	tt.t.Helper()
	if !tt.indexSet {
		tt.t.Fatalf(str, args...)
	}
	indexPart := fmt.Sprintf("[%d] ", tt.index)
	tt.t.Fatalf(indexPart+str, args...)
}

// CheckEqual checks if two values are deeply equal and calls t.Fatalf if not
func (tt *tester) CheckEqual(expected interface{}, got interface{}) {
	nc := numericCompare(expected, got)
	if !(nc == 0 || nc == -2 && reflect.DeepEqual(expected, got)) {
		tt.t.Helper()
		tt.unequalValues(expected, got)
	}
}

// CheckNumericGreater checks if got value is greater than expected
func (tt *tester) CheckNumericGreater(expected interface{}, got interface{}) {
	if numericCompare(expected, got) != 1 {
		tt.t.Helper()
		tt.unequalValues(expected, got)
	}
}

// CheckNumericLess checks if got value is less than expected
func (tt *tester) CheckNumericLess(expected interface{}, got interface{}) {
	if numericCompare(expected, got) != -1 {
		tt.t.Helper()
		tt.unequalValues(expected, got)
	}
}

// CheckEqualAndNoError checks there is no error, and that two values are deeply equal and calls t.Fatalf if not
func (tt *tester) CheckEqualAndNoError(expected interface{}, got interface{}, gotError error) {
	tt.t.Helper()
	tt.CheckNotError(gotError)
	if !reflect.DeepEqual(expected, got) {
		tt.unequalValues(expected, got)
	}
}

// CheckNil checks if value is nil
func (tt *tester) CheckNil(got interface{}) {
	rf := reflect.ValueOf(got)
	if rf.IsValid() && !rf.IsNil() {
		tt.t.Helper()
		tt.Fatalf("Expected: nil, got %v", got)
	}
}

// CheckNotNil checks if value is not nil
func (tt *tester) CheckNotNil(got interface{}) {
	rf := reflect.ValueOf(got)
	if !rf.IsValid() || rf.IsNil() {
		tt.t.Helper()
		tt.Fatalf("Expected: not nil, got nil")
	}
}

// CheckError checks if there is an error
func (tt *tester) CheckError(got error) {
	if got == nil {
		tt.t.Helper()
		tt.Fatalf("Expected: error, got %v", got)
	}
}

// CheckNotError checks if value is not nil
func (tt *tester) CheckNotError(got error) {
	if got != nil {
		tt.t.Helper()
		tt.Fatalf("Expected: no error, got %v", got)
	}
}

// CheckTrue checks if value is true
func (tt *tester) CheckTrue(got bool) {
	if !got {
		tt.t.Helper()
		tt.Fatalf("Expected: true, got %v", got)
	}
}

// CheckFalse checks if value is false
func (tt *tester) CheckFalse(got bool) {
	if got {
		tt.t.Helper()
		tt.Fatalf("Expected: false, got %v", got)
	}
}

// CheckAfter checks if actual value is after the expected value with optional added duration
func (tt *tester) CheckAfter(expected, got time.Time, add ...time.Duration) {
	for _, d := range add {
		expected = expected.Add(d)
	}
	if !got.After(expected) {
		tt.t.Helper()
		tt.Fatalf("Expected: time after %v, got %v (diff %v)", expected, got, got.Sub(expected))
	}
}

// CheckAfterOrEqual checks if actual value is equal or after the expected value with optional added duration
func (tt *tester) CheckAfterOrEqual(expected, got time.Time, add ...time.Duration) {
	for _, d := range add {
		expected = expected.Add(d)
	}
	if expected.Equal(got) {
		return
	}
	if !got.After(expected) {
		tt.t.Helper()
		tt.Fatalf("Expected: time after %v, got %v (diff %v)", expected, got, got.Sub(expected))
	}
}

// CheckBefore checks if actual value is before the expected value with optional added duration
func (tt *tester) CheckBefore(expected, got time.Time, add ...time.Duration) {
	for _, d := range add {
		expected = expected.Add(d)
	}
	if !got.Before(expected) {
		tt.t.Helper()
		tt.Fatalf("Expected: time before %v, got %v (diff %v)", expected, got, got.Sub(expected))
	}
}

// CheckBeforeOrEqual checks if actual value is before the expected value with optional added duration
func (tt *tester) CheckBeforeOrEqual(expected, got time.Time, add ...time.Duration) {
	for _, d := range add {
		expected = expected.Add(d)
	}
	if expected.Equal(got) {
		return
	}
	if !got.Before(expected) {
		tt.t.Helper()
		tt.Fatalf("Expected: time before %v, got %v (diff %v)", expected, got, got.Sub(expected))
	}
}

// CheckMatches checks expected regular expression is matched by the given string and calls t.Fatalf if not
//
// The expected regular expression can be either a *regexp.Regexp or a string that represents a valid regexp
func (tt *tester) CheckMatches(expected interface{}, got string) {
	var rx *regexp.Regexp
	switch expected := expected.(type) {
	case *regexp.Regexp:
		rx = expected
	case string:
		var err error
		rx, err = regexp.Compile(expected)
		if err != nil {
			tt.t.Helper()
			tt.Fatalf("CheckMatches: illegal regexp %q", expected)
		}
	default:
		tt.t.Helper()
		tt.Fatalf("CheckMatches: first argument must be a regexp or a string, got %T %v", expected, expected)
	}
	if !rx.MatchString(got) {
		tt.t.Helper()
		tt.Fatalf("Expected match for %q, got %s", rx.String(), got)
	}
}

// CheckTruef takes a predicate (outcome of a test) and if it is false calls tester t Failf
func (tt *tester) CheckTruef(predicate bool, fmt string, args ...interface{}) {
	if predicate {
		return
	}
	tt.t.Fatalf(fmt, args...)
}
