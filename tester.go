package testutils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
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
	CheckNotEqual(expected interface{}, got interface{})
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
	CheckStringSlicesEqual(expected, got []string)
	CheckTextEqual(expected, got string)
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
	tt.Fatalf("Expected Equal: %T %v, got %T %v", e, e, g, g)
}
func (tt *tester) equalValues(e, g interface{}) {
	tt.t.Helper()
	tt.Fatalf("Expected Noti Equal: %T %v, got %T %v", e, e, g, g)
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

// CheckNotEqual checks if two values are deeply equal and calls t.Fatalf if not
func (tt *tester) CheckNotEqual(expected interface{}, got interface{}) {
	nc := numericCompare(expected, got)
	if nc == 0 || nc == -2 && reflect.DeepEqual(expected, got) {
		tt.t.Helper()
		tt.equalValues(expected, got)
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
	tt.t.Helper()
	tt.t.Fatalf(fmt, args...)
}

// CheckStringSlicesEqual
func (tt *tester) CheckStringSlicesEqual(expected, got []string) {
	diff, ok := produceDiff(expected, got)
	if !ok {
		tt.t.Helper()
		tt.t.Fatalf("slices not equal - see diff:\n%s", diff)
	}
}

// Produces expected and actual interleaved with a not if the are equal or not. Returns ok if there is no diff
// and a each index below each other output for easy human comparison of mismatched result.
func produceDiff(expected, got []string) (diff string, ok bool) {
	cmpE := expected
	cmpG := got
	lE := len(expected)
	lG := len(got)
	if lE < lG {
		cmpE = make([]string, lG)
		copy(cmpE, expected)
	}
	if lE > lG {
		cmpG = make([]string, lE)
		copy(cmpG, got)
	}
	isDiff := false
	var result []string
	ok = true
	badCount := 0
	for i, e := range cmpE {
		isDiff = (e != cmpG[i])
		markerE := " = "
		markerG := " = "
		switch {
		case isDiff && lE < lG && i >= lE:
			markerE = "-! "
			markerG = " !+"
		case isDiff && lE > lG && i >= lG:
			markerE = "+! "
			markerG = " !-"
		case isDiff:
			markerE = " ! "
			markerG = " ! "
		}
		if isDiff {
			ok = false
		}

		// add expected and then got
		if !isDiff {
			result = append(result, fmt.Sprintf("%s eg[%d] `%s`", markerE, i, e))
			badCount = 0
		} else {
			result = append(result, fmt.Sprintf("%s  e[%d] `%s`", markerE, i, e))
			result = append(result, fmt.Sprintf("%s  g[%d] `%s`", markerG, i, cmpG[i]))
			badCount++
			if badCount > 2 {
				result = append(result, "... stopping after 2 unequal lines")
				break
			}
		}
	}
	return strings.Join(result, "\n"), ok
}

// CheckTextEqual behaves like CheckEqual in general, but in addition to just failing
// a color coded diff will be produced in the error message making it easier to see where the
// difference is (when run in a terminal window).
func (tt *tester) CheckTextEqual(expected, got string) {
	if expected != got {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, got, false)
		pretty := dmp.DiffPrettyText(diffs)
		tt.t.Fatalf("strings not equal - see diff:\n%s", pretty)
	}
}
