package coverage

import (
	"io"
	"strings"
	"testing"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func mockGoTest(data string) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		return nopCloser{strings.NewReader(data)}, nil
	}
}

func TestRunNoUncovered(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:1.1,2.1 1 1\n")

	var sb strings.Builder
	if err := Run(nil, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunUncovered(t *testing.T) {
	old := runGoTest
	defer func() { runGoTest = old }()

	runGoTest = mockGoTest("mode: set\ngithub.com/app/main.go:10.1,12.2 2 0\n")

	var sb strings.Builder
	if err := Run(nil, &sb); err != ErrUncovered {
		t.Fatalf("expected ErrUncovered, got %v", err)
	}
}

func TestRunVersion(t *testing.T) {
	var sb strings.Builder
	if err := Run([]string{"--version"}, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if sb.Len() == 0 {
		t.Fatal("expected version output")
	}
}

func TestRunHelp(t *testing.T) {
	var sb strings.Builder
	if err := Run([]string{"--help"}, &sb); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if sb.Len() == 0 {
		t.Fatal("expected help output")
	}
}
