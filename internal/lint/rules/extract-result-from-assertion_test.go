package rules_test

import (
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"coderaiser/go-coverage/internal/lint/rules"
)

// checkRule parses src, runs Check, and returns the number of violations.
func checkRule(t *testing.T, src string) int {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "t.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	r := &rules.ExtractResultFromAssertion{}
	return len(r.Check(file, fset))
}

// fixRule parses src, runs Fix, and returns the formatted result.
func fixRule(t *testing.T, src string) string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "t.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	r := &rules.ExtractResultFromAssertion{}
	r.Fix(file, fset)

	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		t.Fatalf("format: %v", err)
	}
	return buf.String()
}

const pkgHeader = `package coverage_test

import tape "github.com/coderaiser/go-tape"

`

func TestExtractResultFromAssertion_Check(t *testing.T) {
	t.Run("flags inline call in first arg", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		t.Equal(someFunc(), "expected")
		t.End()
	})
}`
		if got := checkRule(t, src); got != 1 {
			t.Fatalf("want 1 violation, got %d", got)
		}
	})

	t.Run("flags inline composite literal in second arg", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		t.DeepEqual(result, []string{"a", "b"})
		t.End()
	})
}`
		if got := checkRule(t, src); got != 1 {
			t.Fatalf("want 1 violation, got %d", got)
		}
	})

	t.Run("flags both args needing extraction", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		t.DeepEqual(someFunc(x), []string{"a"})
		t.End()
	})
}`
		if got := checkRule(t, src); got != 1 {
			t.Fatalf("want 1 violation, got %d", got)
		}
	})

	t.Run("no violation when result already declared", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		result := someFunc()
		t.Equal(result, "expected")
		t.End()
	})
}`
		if got := checkRule(t, src); got != 0 {
			t.Fatalf("want 0 violations, got %d", got)
		}
	})

	t.Run("no violation when both vars already declared", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		result := someFunc()
		expected := []string{"a"}
		t.DeepEqual(result, expected)
		t.End()
	})
}`
		if got := checkRule(t, src); got != 0 {
			t.Fatalf("want 0 violations, got %d", got)
		}
	})

	t.Run("no violation for simple identifier args", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		result := someFunc()
		t.Equal(result, want)
		t.End()
	})
}`
		if got := checkRule(t, src); got != 0 {
			t.Fatalf("want 0 violations, got %d", got)
		}
	})

	t.Run("no violation for t.End", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		result := someFunc()
		t.Equal(result, want)
		t.End()
	})
}`
		if got := checkRule(t, src); got != 0 {
			t.Fatalf("want 0 violations, got %d", got)
		}
	})

	t.Run("no violation for unrelated receiver", func(t *testing.T) {
		src := pkgHeader + `
func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		other.Equal(someFunc(), "x")
		t.End()
	})
}`
		if got := checkRule(t, src); got != 0 {
			t.Fatalf("want 0 violations, got %d", got)
		}
	})
}

func TestExtractResultFromAssertion_Fix(t *testing.T) {
	t.Run("extracts call from first arg", func(t *testing.T) {
		src := pkgHeader + `func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		t.Equal(someFunc(), "expected")
		t.End()
	})
}`
		got := fixRule(t, src)
		if !strings.Contains(got, "result := someFunc()") {
			t.Fatalf("expected result extraction, got:\n%s", got)
		}
		if !strings.Contains(got, "t.Equal(result,") {
			t.Fatalf("expected assertion to use result, got:\n%s", got)
		}
	})

	t.Run("extracts composite literal from second arg", func(t *testing.T) {
		src := pkgHeader + `func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		t.DeepEqual(result, []string{"a", "b"})
		t.End()
	})
}`
		got := fixRule(t, src)
		if !strings.Contains(got, `expected := []string{"a", "b"}`) {
			t.Fatalf("expected extraction of composite lit, got:\n%s", got)
		}
		if !strings.Contains(got, "t.DeepEqual(result, expected)") {
			t.Fatalf("expected assertion to use expected, got:\n%s", got)
		}
	})

	t.Run("extracts both args", func(t *testing.T) {
		src := pkgHeader + `func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		t.DeepEqual(someFunc(x), []string{"a"})
		t.End()
	})
}`
		got := fixRule(t, src)
		if !strings.Contains(got, "result := someFunc(x)") {
			t.Fatalf("expected result extraction, got:\n%s", got)
		}
		if !strings.Contains(got, `expected := []string{"a"}`) {
			t.Fatalf("expected expected extraction, got:\n%s", got)
		}
		if !strings.Contains(got, "t.DeepEqual(result, expected)") {
			t.Fatalf("expected assertion rewrite, got:\n%s", got)
		}
	})

	t.Run("does not touch already-extracted assertions", func(t *testing.T) {
		src := pkgHeader + `func TestFoo(t *testing.T) {
	tape.Test(t, "foo", func(t *tape.T) {
		result := someFunc()
		expected := []string{"a"}
		t.DeepEqual(result, expected)
		t.End()
	})
}`
		got := fixRule(t, src)
		// Should appear exactly once — not duplicated.
		if count := strings.Count(got, "result :="); count != 1 {
			t.Fatalf("expected 1 result assignment, got %d:\n%s", count, got)
		}
		if count := strings.Count(got, "expected :="); count != 1 {
			t.Fatalf("expected 1 expected assignment, got %d:\n%s", count, got)
		}
	})

	t.Run("real-world: MergeBlocks test", func(t *testing.T) {
		src := pkgHeader + `func TestMergeBlocks(t *testing.T) {
	tape.Test(t, "coverage: MergeBlocks merges overlapping same-file blocks", func(t *tape.T) {
		result := MergeBlocks([]Block{
			{File: "a.go", Start: 10, End: 10},
			{File: "a.go", Start: 10, End: 12},
		})
		t.DeepEqual(result, []Block{
			{File: "a.go", Start: 10, End: 15},
		})
		t.End()
	})
}`
		got := fixRule(t, src)
		// result already declared — should not be re-extracted
		if count := strings.Count(got, "result :="); count != 1 {
			t.Fatalf("result extracted twice, got %d:\n%s", count, got)
		}
		// the []Block literal in DeepEqual should be extracted to expected
		if !strings.Contains(got, "expected :=") {
			t.Fatalf("expected 'expected' extraction, got:\n%s", got)
		}
		if !strings.Contains(got, "t.DeepEqual(result, expected") {
			t.Fatalf("expected assertion rewrite, got:\n%s", got)
		}
	})
}
