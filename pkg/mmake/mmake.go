package mmake

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aakarim/mmake/pkg/mmake/completion"
	"github.com/aakarim/mmake/pkg/mmake/workspace"
)

type MMake struct {
}

func New() *MMake {
	return &MMake{}
}

func (m *MMake) Run(ctx context.Context, inputPath string, args ...string) error {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := m.Init(ctx); err != nil {
			panic(err)
		}
		return nil
	}
	workspacePath, err := workspace.FindWorkspaceFile(ctx, inputPath)
	if err != nil {
		return err
	}
	ws := workspace.New(filepath.Dir(workspacePath))

	if err := ws.Init(ctx); err != nil {
		return err
	}

	// if args[1] starts with '//' then it's a target
	if len(args) > 1 && args[1][:2] == "//" {
		target := args[1]
		if err := ws.RunTarget(ctx, target); err != nil {
			return err
		}
		return nil
	}

	// if args[1] is "completion" then we want to print the completion script
	if len(args) > 1 && args[1] == "completion" {
		fmt.Println(completion.GetCompletionScript(workspacePath))
		return nil
	}

	// if args[1] is "query" then we want to query the Workspace
	if len(args) > 1 && args[1] == "query" {
		qu := workspace.NewQuery(ws)
		if err := qu.Update(ctx); err != nil {
			return err
		}
		if len(args) < 3 {
			return fmt.Errorf("query required")
		}

		files, err := qu.QueryFilesByPrefix(ctx, args[2])
		if err != nil {
			return err
		}
		// print the files as columns with file and description
		for _, file := range files {
			packageName, err := qu.GetPackageFromFile(file.Path)
			if err != nil {
				return err
			}
			fmt.Printf("%s\t%s\n", packageName, file.Description)
		}

		return nil
	}

	// if args[1] is "clean" then we want to clean the target dir
	if len(args) > 1 && args[1] == "clean" {
		if err := ws.Clean(ctx, args[2]); err != nil {
			return err
		}
		return nil
	}

	return ErrNoCommand
}

// Init creates a new WORKSPACE.mmake file in the current directory
// TODO: move this into the workspace package
func (m *MMake) Init(ctx context.Context) error {
	_, err := os.Create("WORKSPACE.mmake")
	if err != nil {
		return err
	}

	return nil
}
