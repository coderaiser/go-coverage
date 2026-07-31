# coverage

CLI tool that runs `go test` and prints uncovered blocks.

```
$ coverage
💪 coverage 100%, good job!
```

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `lines` | Output format: `lines`, `code-frame`, `json-lines` |

## Quick start

```sh
# Show uncovered blocks
coverage

# Show with source context
coverage -f code-frame

# Machine-readable JSON output
coverage -f json-lines
```

Set `COLOR=0` to disable ANSI colours (useful in CI or when piping output).

```sh
COLOR=0 coverage
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
