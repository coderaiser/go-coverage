package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/coderaiser/go-coverage"
	"github.com/coderaiser/go-coverage/internal/runner"
)

func main() {
	if err := runner.Run(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, coverage.ErrUncovered) || errors.Is(err, runner.ErrTestFailed) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
