package main

import (
	"context"
	"fmt"
	"os"

	"github.com/zack-nova/orbit/cmd/harness/cli"
)

func main() {
	if err := cli.Execute(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
