package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/aakarim/mmake/pkg/mmake"
	"github.com/aakarim/mmake/pkg/mmake/workspace"
)

var workspacePath = flag.String("w", "", "path to workspace")

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	flag.Parse()

	mmake := mmake.New()

	if err := mmake.Run(ctx, *workspacePath, os.Args...); err != nil {
		var cmdErr *workspace.ErrCommand
		if errors.As(err, &cmdErr) {
			// if the error is a command error, then we want to exit with the exit code
			// and not print the stack trace since it's a user error.
			os.Exit(1)
			return
		}
		fmt.Println("error:", err)
	}
}
