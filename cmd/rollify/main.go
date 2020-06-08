package main

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Run runs the main application.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) error {

	return nil
}

func main() {
	ctx := context.Background()

	err := Run(ctx, os.Args, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
