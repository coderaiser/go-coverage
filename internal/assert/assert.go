package assert

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
)

var (
	mu    sync.Mutex
	count = make(map[*testing.T]int)
)

func One(t *testing.T) {
	t.Helper()

	mu.Lock()
	count[t] = 0
	mu.Unlock()
}

func hit(t *testing.T) {
	t.Helper()

	mu.Lock()
	defer mu.Unlock()

	count[t]++

	if count[t] > 1 {
		_, file, line, ok := runtime.Caller(2)

		if ok {
			t.Fatalf(
				"too many assertions: got %d, expected 1\nat %s:%d",
				count[t],
				file,
				line,
			)
		}

		t.Fatalf(
			"too many assertions: got %d, expected 1",
			count[t],
		)
	}
}

func Equal(t *testing.T, want, got any) {
	t.Helper()
	hit(t)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("\nwant: %#v\n got: %#v", want, got)
	}
}

func Contains(t *testing.T, s, sub string) {
	t.Helper()
	hit(t)

	if !strings.Contains(s, sub) {
		t.Errorf("%q does not contain %q", s, sub)
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()
	hit(t)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Error(t *testing.T, err error) {
	t.Helper()
	hit(t)

	if err == nil {
		t.Fatal(fmt.Errorf("expected an error, got nil"))
	}
}
func Ok(t *testing.T, v bool) {
	t.Helper()
	hit(t)

	if !v {
		t.Errorf("expected true, got false")
	}
}

func NotOk(t *testing.T, v bool) {
	t.Helper()
	hit(t)

	if v {
		t.Errorf("expected false, got true")
	}
}
