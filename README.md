# Coverage [![License][LicenseIMGURL]][LicenseURL] [![Build Status][BuildStatusIMGURL]][BuildStatusURL] [![Coverage Status][CoverageIMGURL]][CoverageURL]

[BuildStatusURL]: https://github.com/coderaiser/go-coverage/actions/workflows/test.yml
[BuildStatusIMGURL]: https://github.com/coderaiser/go-coverage/actions/workflows/test.yml/badge.svg
[LicenseURL]: https://tldrlegal.com/license/mit-license "MIT License"
[LicenseIMGURL]: https://img.shields.io/badge/license-MIT-317BF9.svg?style=flat
[CoverageURL]: https://coveralls.io/github/coderaiser/go-tape?branch=master
[CoverageIMGURL]: https://coveralls.io/repos/coderaiser/go-tape/badge.svg?branch=master&service=github

CLI tool that runs `go test` and prints uncovered blocks.

```
$ coverage
💪 coverage 100%, good job!
```

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `lines` | Output format: `lines`, `codeframe`, `json-lines` |
| `-r, --report [file]` | `coverage.lcov` | Write lcov report (for coveralls, codecov, genhtml) |

## Quick start

```sh
# Show uncovered blocks
coverage

# Show with source context
coverage -f codeframe

# Machine-readable JSON output
coverage -f json-lines

# Write lcov report (default: coverage.lcov)
coverage -r

# Write lcov report to a custom path
coverage -r lcov.info

# Check coverage and produce lcov report in one pass
coverage -r coverage.lcov -f codeframe
```

Set `COLOR=0` to disable ANSI colours (useful in CI or when piping output).

```sh
COLOR=0 coverage
```

## Configuration

Place a `.coverage.toml` in your project root to exclude files from coverage checks and the lcov report:

```toml
[exclude]
files = ["**/testdata", "**/mock"]
```

Excluded files are omitted from both the terminal output and the lcov report.

## CI usage

`-r` lets you check coverage and produce a report in a single `go test` run — no need to run tests twice:

```yaml
# GitHub Actions example
- run: coverage -r coverage.lcov
- uses: coverallsapp/github-action@v2
  with:
    file: coverage.lcov
```

## Install

The simplest possible way is with [palabra](https://github.com/coderaiser/palabra):

```sh
palabra i coverage
```

Or just grab binaries from [releases](https://github.com/coderaiser/go-coverage/releases)

## Running tests

```sh
task test
```
