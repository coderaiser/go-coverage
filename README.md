# coverage

CLI tool that reads a Go `coverage.out` profile and prints coverage blocks.
Optionally shows a code frame (source lines) around each coverage block.

```
coverage [-f coverage.out] [--code-frame]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-f / --f` | `coverage.out` | Path to the coverage profile |
| `--code-frame / -code-frame` | off | Print source lines around each block |

## Quick start

```sh
# 1. Generate a coverage profile for your own project
go test -coverprofile=coverage.out ./...

# 2. Show coverage blocks
coverage

# 3. Show coverage blocks with source context
coverage --code-frame

# 4. Point at a custom profile
coverage -f /tmp/myproject.out --code-frame
```

Set `COLOR=0` to disable ANSI colours (useful in CI or when piping output).

```sh
COLOR=0 coverage -f coverage.out
```

## Install

**From source (requires Go 1.22+)**

```sh
git clone https://github.com/your-org/coverage.git
cd coverage
go install ./cmd/coverage
```

**Pre-built binaries**

Download the binary for your platform from the
[Releases](https://github.com/your-org/coverage/releases) page:

| OS | Arch | File |
|----|------|------|
| Linux | amd64 | `coverage-linux-amd64` |
| Linux | arm64 | `coverage-linux-arm64` |
| macOS | amd64 | `coverage-darwin-amd64` |
| macOS | arm64 | `coverage-darwin-arm64` |
| Windows | amd64 | `coverage-windows-amd64.exe` |
| Windows | arm64 | `coverage-windows-arm64.exe` |

```sh
palabra i coverage
```

## Running tests

```sh
make test
```

### Coverage report

`run.go` contains only CLI glue (`os.Open`, `os.Args`, `flag.Parse`) and is
excluded from coverage via `-coverpkg` scoping.  
All library functions in `coverage.go` are covered at **100 %**.

```sh
make coverage        # prints per-function table + opens HTML report
```

Expected output:

```
coderaiser/go-coverage  ColorEnabled     100.0%
coderaiser/go-coverage  ParseCoverage    100.0%
coderaiser/go-coverage  ReadLines        100.0%
coderaiser/go-coverage  FormatBlock      100.0%
coderaiser/go-coverage  total:           100.0%
```

## Project layout

```
.
├── coverage.go          # Library: ParseCoverage, ReadLines, FormatBlock, ColorEnabled
├── run.go                # CLI glue: Run() — excluded from coverage
├── coverage_test.go     # Unit tests (package coverage_test)
├── cmd/
│   └── coverage/
│       └── main.go       # Binary entry point
├── internal/
│   └── assert/
│       └── assert.go     # Minimal test helpers (testify-compatible API)
├── Makefile
└── .github/
    └── workflows/
        └── release.yml   # Build & publish binaries on git tag
```
