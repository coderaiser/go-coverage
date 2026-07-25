// Package assert provides minimal test helpers matching the testify/assert API.
package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Equal(t *testing.T, want, got any) {
	t.Helper()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("\nwant: %#v\n got: %#v", want, got)
	}
}

func Contains(t *testing.T, s, sub string) {
	t.Helper()

	if !strings.Contains(s, sub) {
		t.Errorf("%q does not contain %q", s, sub)
	}
}

func NotOk(t *testing.T, v bool) {
	t.Helper()

	if v {
		t.Errorf("expected false, got true")
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Error(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal(fmt.Errorf("expected an error, got nil"))
	}
}

func Ok(t *testing.T, v bool) {
	t.Helper()

	if !v {
		t.Errorf("expected true, got false")
	}
}
