package main

import (
	"errors"
	"fmt"
	"os"

	"coderaiser/go-coverage"
)

func main() {
	if err := coverage.Run(os.Args[1:], os.Stdout); err != nil {
		if errors.Is(err, coverage.ErrUncovered) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
