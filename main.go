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
var help = flag.Bool("h", false, "print help")

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
		return
	}

	mm := mmake.New()

	if err := mm.Run(ctx, *workspacePath, os.Args...); err != nil {
		var cmdErr *workspace.ErrCommand
		if errors.As(err, &cmdErr) {
			// if the error is a command error, then we want to exit with the exit code
			// and not print the stack trace since it's a user error.
			os.Exit(1)
			return
		}
		if errors.Is(err, mmake.ErrNoCommand) {
			printUsage()
			os.Exit(1)
			return
		}
		fmt.Println("error:", err)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s [target | command] [target | command]:\n", os.Args[0])

	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  init\t\tInitialize a new workspace\n")
	fmt.Fprintf(os.Stderr, "  completion\tPrint the completion script\n")
	fmt.Fprintf(os.Stderr, "  clean\tRemove the package's build artifacts folder\n")
	fmt.Fprintf(os.Stderr, "  info\tRetrieve information about target\n")
	fmt.Fprintf(os.Stderr, "  //[path]:[target]\tRun a specific target\n")
	fmt.Fprintf(os.Stderr, "\n")
}
