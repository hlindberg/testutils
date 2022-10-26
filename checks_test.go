package testutils

import (
	"io"
	"testing"
)

func ensureFailed(t *testing.T, f func(t *testing.T)) {
	tt := testing.T{}
	x := make(chan bool, 1)
	go func() {
		defer func() { x <- true }() // GoExit runs all deferred calls
		f(&tt)
	}()
	<-x
	if !tt.Failed() {
		t.Fail()
	}
}

func ensureNotFailed(t *testing.T, f func(t *testing.T)) {
	tt := testing.T{}
	x := make(chan bool, 1)
	go func() {
		defer func() { x <- true }() // GoExit runs all deferred calls
		f(&tt)
	}()
	<-x
	if tt.Failed() {
		t.Fail()
	}
}

func TestCheckEqual(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckEqual("a", "b", ft)
	})
}

func TestCheckNil(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckNil([]byte{0}, ft)
	})
}

func TestCheckNotNil(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckNotNil(nil, ft)
	})
}

func TestCheckError(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckError(nil, ft)
	})
}

func TestCheckNotError(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckNotError(io.ErrUnexpectedEOF, ft)
	})
}

func TestCheckTrue(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckTrue(false, ft)
	})
}

func TestCheckFalse(t *testing.T) {
	ensureFailed(t, func(ft *testing.T) {
		CheckFalse(true, ft)
	})
}

func TestCheckEqualElements(t *testing.T) {
	expected := []string{"a", "b", "c"}
	got := []string{"a", "b", "c"}
	ensureNotFailed(t, func(ft *testing.T) {
		CheckEqualElements(expected, got, ft)
	})

	got = []string{"c", "b", "a"}
	ensureNotFailed(t, func(ft *testing.T) {
		CheckEqualElements(expected, got, ft)
	})

	got = []string{"a"}
	ensureFailed(t, func(ft *testing.T) {
		CheckEqualElements(expected, got, ft)
	})
}

func Test_valuesEqual(t *testing.T) {
	if !valuesEqual(1, 1) {
		t.Fail()
	}
	if valuesEqual("1", 1) {
		t.Fail()
	}
	if !valuesEqual("a", "a") {
		t.Fail()
	}
	if !valuesEqual(false, false) {
		t.Fail()
	}
	if valuesEqual(false, true) {
		t.Fail()
	}
	if !valuesEqual([]string{"a"}, []string{"a"}) {
		t.Fail()
	}
	if valuesEqual([]string{"a", "b"}, []string{"b", "a"}) {
		t.Fail()
	}
}
