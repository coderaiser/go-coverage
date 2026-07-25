package main

import (
	"fmt"
	"os"

	"coderaiser/go-coverage"
)

func main() {
	if err := coverage.Run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
