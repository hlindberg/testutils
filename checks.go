// Package testutils contains convenient testing checkers that compare a produced
// value against an expected value (or condition).
// There are value checks like `CheckEqual(expected, produced, t)“, and
// checks that should run deferred like `defer ShouldPanic(t)`.
package testutils

import (
	"bytes"
	"io"
	"math"
	"os"
	"reflect"
	"regexp"
	"testing"
)

func unequalValues(e, g interface{}, t *testing.T) {
	t.Helper()
	t.Fatalf("Expected equal: %T %v, got %T %v", e, e, g, g)
}
func equalValues(e, g interface{}, t *testing.T) {
	t.Helper()
	t.Fatalf("Expected not equal: %T %v, got %T %v", e, e, g, g)
}

// CheckEqual checks if two values are deeply equal and calls t.Fatalf if not
func CheckEqual(expected interface{}, got interface{}, t *testing.T) {
	if !valuesEqual(expected, got) {
		t.Helper()
		unequalValues(expected, got, t)
	}
}

// CheckNotEqual checks if two values are deeply equal and calls t.Fatalf if not
func CheckNotEqual(expected interface{}, got interface{}, t *testing.T) {
	if valuesEqual(expected, got) {
		t.Helper()
		equalValues(expected, got, t)
	}
}

// CheckMatches checks expected regular expression is matched by the given string and calls t.Fatalf if not
//
// The expected regular expression can be either a *regexp.Regexp or a string that represents a valid regexp
func CheckMatches(expected interface{}, got string, t *testing.T) {
	var rx *regexp.Regexp
	switch expected := expected.(type) {
	case *regexp.Regexp:
		rx = expected
	case string:
		var err error
		rx, err = regexp.Compile(expected)
		if err != nil {
			t.Fatalf("CheckMatches: illegal regexp %q", expected)
		}
	default:
		t.Fatalf("CheckMatches: first argument must be a regexp or a string, got %T %v", expected, expected)
	}
	if !rx.MatchString(got) {
		t.Helper()
		t.Fatalf("Expected match for %q, got %s", rx.String(), got)
	}
}

// func valuesEqual(a interface{}, b interface{}) bool {
// 	nc := numericCompare(a, b)
// 	return nc == 0 || nc == -2 && reflect.DeepEqual(a, b)
// }

func valuesEqual(a interface{}, b interface{}) bool {
	nc := numericCompare(a, b)
	var ok bool
	if nc == -2 {
		// Not numerically equal
		// Since DeepEqual does not return true for some cases of reflected values. (Not sure why).
		sa, ok1 := AsInterface(a)
		sb, ok2 := AsInterface(b)
		if ok1 && ok2 && sa == sb {
			return true
		}
		ok = reflect.DeepEqual(a, b)
	}
	return nc == 0 || nc == -2 && ok
}

// CheckEqualAndNoError checks there is no error, and that two values are deeply equal and calls t.Fatalf if not
func CheckEqualAndNoError(expected interface{}, got interface{}, gotError error, t *testing.T) {
	t.Helper()
	CheckNotError(gotError, t)
	if !reflect.DeepEqual(expected, got) {
		unequalValues(expected, got, t)
	}
}

// CheckContainsElements checks if one slice contains all elements of another slice irrespective of order and uniqueness.
func CheckContainsElements(expected interface{}, got interface{}, t *testing.T) {
	if sliceContains(got, expected, false) {
		t.Helper()
		t.Fatalf("Slice %v does not contain all elements in %v", got, expected)
	}
}

// CheckEqualElements checks if two slices contains the exact same set of elements irrespective of order and uniqueness.
func CheckEqualElements(expected interface{}, got interface{}, t *testing.T) {
	if !sliceContains(got, expected, true) {
		t.Helper()
		t.Fatalf("Elements of slice %v and %v differ", expected, got)
	}
}

// sliceContainsAll returns true if a contains all elements in b, irrespective of order. Each element is
// matched exactly once. If checkSize is given a and b must have the same size.
func sliceContains(a, b interface{}, checkSize bool) bool {
	va := reflect.ValueOf(a)
	if va.Kind() != reflect.Slice {
		return false
	}
	vb := reflect.ValueOf(b)
	if vb.Kind() != reflect.Slice {
		return false
	}
	tb := vb.Len()
	ta := va.Len()

	if checkSize && ta != tb {
		return false
	}
	// any set contains an empty set
	if tb == 0 {
		return true
	}
	// an empty set cannot contain a non empty set
	if ta == 0 {
		return false
	}
	// a smaller set cannot contain all from a larger set
	if ta < tb {
		return false
	}
	ma := make([]bool, ta)

nextB:
	for ib := 0; ib < tb; ib++ {
		eb := vb.Index(ib)
		for ia := 0; ia < ta; ia++ {
			if ma[ia] {
				continue
			}
			if valuesEqual(eb, va.Index(ia)) {
				ma[ia] = true
				continue nextB
			}
		}
		return false
	}
	return true
}

// CheckNil checks if value is nil
func CheckNil(got interface{}, t *testing.T) {
	rf := reflect.ValueOf(got)
	if rf.IsValid() && !rf.IsNil() {
		t.Helper()
		t.Fatalf("Expected: nil, got %v", got)
	}
}

// CheckNotNil checks if value is not nil
func CheckNotNil(got interface{}, t *testing.T) {
	rf := reflect.ValueOf(got)
	if !rf.IsValid() || rf.IsNil() {
		t.Helper()
		t.Fatalf("Expected: not nil, got nil")
	}
}

// CheckError checks if there is an error
func CheckError(got interface{}, t *testing.T) {
	_, ok := got.(error)
	if !ok {
		t.Helper()
		t.Fatalf("Expected: error, got %v", got)
	}
}

// CheckNotError checks if value is not an error
func CheckNotError(got interface{}, t *testing.T) {
	err, ok := got.(error)
	if ok {
		t.Helper()
		t.Fatalf("Expected: no error, got %q", err.Error())
	}
}

// CheckNumericGreater checks if second value is greater than first. Comparisons are made regardless of
// bit size and an integer is equal to a float if casting it to a float makes it equal.
func CheckNumericGreater(expected interface{}, got interface{}, t *testing.T) {
	if numericCompare(expected, got) != 1 {
		t.Helper()
		t.Fatalf("Expected: %T %v greater than %T %v", expected, expected, got, got)
	}
}

// CheckNumericLess checks if second value is less than first. Comparisons are made regardless of
// bit size and an integer is equal to a float if casting it to a float makes it equal.
func CheckNumericLess(expected interface{}, got interface{}, t *testing.T) {
	if numericCompare(expected, got) != -1 {
		t.Helper()
		t.Fatalf("Expected: %T %v less than %T %v", expected, expected, got, got)
	}
}

// numericCompare checks if two numeric values are equal. Comparisons are made regardless of
// bit size and an integer is equal to a float if casting it to a float makes it equal.
// Return 0 if numerically equal, 1 if got is greater than expected, and -1 if less.
// In case they are not numeric values a value of -2 is returned.
func numericCompare(expected interface{}, got interface{}) int {
	if ei, ok := AsInteger(expected); ok {
		if gi, ok := AsInteger(got); ok {
			if ei == gi {
				return 0
			} else if gi > ei {
				return 1
			}
			return -1
		} else if gf, ok := AsFloat(got); ok {
			if float64(ei) == gf {
				return 0
			} else if gf > float64(ei) {
				return 1
			}
			return -1
		}
	} else if ef, ok := AsFloat(expected); ok {
		if gf, ok := AsFloat(got); ok {
			if ef == gf {
				return 0
			} else if gf > ef {
				return 1
			}
			return -1
		}
		// No need to test if "got" is an integer since we know that "expected" isn't
	} else if eu, ok := expected.(uint64); ok {
		if gu, ok := expected.(uint64); ok && eu == gu {
			return 0
		} else if gu > eu {
			return 1
		}
		return -1
		// No need to test if "got" is an integer since we know that "expected" isn't
	}
	return -2
}

// AsInteger returns the argument as a signed 64 bit integer and true if the argument
// is an integer that fits into that type. Otherwise it returns 0, false
func AsInteger(v interface{}) (int64, bool) {
	ok := true
	var rv int64
	switch et := v.(type) {
	case int8:
		rv = int64(et)
	case int16:
		rv = int64(et)
	case int32:
		rv = int64(et)
	case int:
		rv = int64(et)
	case int64:
		rv = et
	case uint8:
		rv = int64(et)
	case uint16:
		rv = int64(et)
	case uint32:
		rv = int64(et)
	case uint:
		if et <= math.MaxInt64 {
			rv = int64(et)
		} else {
			ok = false
		}
	case uint64:
		if et <= math.MaxInt64 {
			rv = int64(et)
		} else {
			ok = false
		}
	default:
		ok = false
	}
	return rv, ok
}

// AsFloat returns the argument as a 64 bit float and true if the argument
// is a float that fits into that type. Otherwise it returns 0, false
func AsFloat(v interface{}) (rv float64, ok bool) {
	ok = true
	switch et := v.(type) {
	case int16:
		rv = float64(et)
	case int32:
		rv = float64(et)
	case int64:
		rv = float64(et)
	case uint16:
		rv = float64(et)
	case uint32:
		rv = float64(et)
	case uint64:
		rv = float64(et)
	case float32:
		rv = float64(et)
	case float64:
		rv = et
	default:
		ok = false
	}
	return
}

// AsInterface returns the argument as an interface{}. This is useful when a value
// may be a reflect.Value and an operation does not work on such, but on the underlying real
// value albeit behind an interface{}.
func AsInterface(v interface{}) (rv interface{}, ok bool) {
	ok = true
	switch et := v.(type) {
	case reflect.Value:
		rv = et.Interface()
	// case string:
	// 	rv = string(et)
	// case bool:
	// 	rv = bool(et)
	default:
		ok = false
	}
	return
}

// CheckTrue checks if value is true
func CheckTrue(got bool, t *testing.T) {
	if !got {
		t.Helper()
		t.Fatalf("Expected: true, got %v", got)
	}
}

// CheckFalse checks if value is false
func CheckFalse(got bool, t *testing.T) {
	if got {
		t.Helper()
		t.Fatalf("Expected: false, got %v", got)
	}
}

const chunkSize = 0x10000

// CheckFilesEqual equals checks if the two files have the exact same contents.
func CheckFilesEqual(file1, file2 string, t *testing.T) {
	t.Helper()
	var fi1, fi2 os.FileInfo
	var err error
	if fi1, err = os.Stat(file1); err != nil {
		t.Fatal(err)
	}

	if fi2, err = os.Stat(file2); err != nil {
		t.Fatal(err)
	}

	if fi1.IsDir() {
		t.Fatalf("%q is a directory", file1)
	}

	if fi2.IsDir() {
		t.Fatalf("%q is a directory", file2)
	}

	sz := fi1.Size()
	if sz != fi2.Size() {
		t.Fatalf("size of file %q (%d), does not match size of %q (%d)", file1, fi1.Size(), file2, fi2.Size())
	}

	var f1, f2 *os.File
	if f1, err = os.Open(file1); err != nil {
		t.Fatal(err)
	}
	defer f1.Close()

	if f2, err = os.Open(file1); err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	if sz > chunkSize {
		sz = chunkSize
	}
	b1 := make([]byte, sz)
	b2 := make([]byte, sz)
	for {
		n1, err1 := f1.Read(b1)
		n2, err2 := f2.Read(b2)

		if err1 == io.EOF && err2 == io.EOF {
			return
		}

		if err1 != nil {
			t.Fatal(err1)
		}

		if err2 != nil {
			t.Fatal(err2)
		}

		if !bytes.Equal(b1[0:n1], b2[0:n2]) {
			t.Fatalf("content of file %q and %q differ", file1, file2)
		}
	}
}

// CheckFileExists checks that given file name is for an existing regular file
func CheckFileExists(filename string, t *testing.T) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		t.Fatalf("file %s does not exist", filename)
	}
	if info.IsDir() {
		t.Fatalf("file %s is a directory, not a file", filename)
	}
}
