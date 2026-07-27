# coverage

CLI tool that reads a Go `coverage.out` profile and prints coverage blocks.
Optionally shows a code frame (source lines) around each coverage block.

```
$ coverage
💪 coverage 100%, good job!
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

The simplest possible way is with [palabra](https://github.com/coderaiser/palabra):

```sh
palabra i coverage
```

Or just grab binaries from [releases](https://github.com/coderaiser/go-coverage/releases)

## Running tests

```sh
task test
```
