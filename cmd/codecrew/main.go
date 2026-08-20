package main

import (
	"fmt"
	"os"

	"github.com/radiusred/gh-codecrew/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "codecrew:", err)
		os.Exit(1)
	}
}
