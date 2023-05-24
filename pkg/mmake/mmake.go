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
	var target string
	var command string

	// if args[1] starts with '//' then it's a target
	if len(args) > 1 && args[1][:2] == "//" {
		target = args[1]
		if len(args) > 2 {
			command = args[2]
		} else {
			command = "run"
		}
	}

	if command == "" && len(args) > 1 {
		command = args[1]
		if len(args) > 2 {
			target = args[2]
		}
	}

	if target == "" && command == "" {
		return ErrNoCommand
	}

	if command == "init" {
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

	if target != "" && workspace.HasCommandToImport(args) {
		if err := ws.Import(ctx, target, args); err != nil {
			return err
		}
		return nil
	}

	if command == "clean" {
		if err := ws.Clean(ctx, target); err != nil {
			return err
		}
		return nil
	}

	if command == "info" {
		info, err := ws.GetInfo(ctx, target)
		if err != nil {
			return err
		}
		fmt.Println(target, "info:")
		fmt.Println(info)
		return nil
	}

	if command == "completion" {
		fmt.Println(completion.GetCompletionScript(workspacePath))
		return nil
	}

	if command == "compgen" {
		prefix := target
		qu := workspace.NewQuery(ws, prefix)
		if err := qu.Update(ctx, 2); err != nil {
			return err
		}

		if len(args) < 3 {
			return fmt.Errorf("query required")
		}

		outputStr, err := qu.GenComp(ctx, prefix)
		if err != nil {
			return err
		}
		fmt.Print(outputStr)
		return nil
	}

	if command == "run" || command == "" {
		if err := ws.RunTarget(ctx, target); err != nil {
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
