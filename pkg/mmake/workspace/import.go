package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrTargetExists = errors.New("target already exists")

// HasCommandToImport returns true if the args contain a command to import a target
// e.g. mmake //services/api:api -- go run ./services/api
// will return true.
func HasCommandToImport(args []string) bool {
	j := strings.Join(args, " ")
	return strings.Contains(j, "--")
}

func GetImportedCommand(args []string) string {
	j := strings.Join(args, " ")
	spl := strings.Split(j, " -- ")
	return spl[1]
}

func (w *Workspace) Import(ctx context.Context, target string, args []string) error {
	// first check if there is a build file at the target
	targetFilePath, err := w.getBuildFile(ctx, target)
	if err != nil && !errors.Is(err, ErrNoMakefileFound) {
		return err
	}
	// if no build file create
	if errors.Is(err, ErrNoMakefileFound) {
		// create a build file
		targetFilePath = filepath.Join(w.rootPath,
			getRelPathFromTarget(target), "Makefile")

	}

	// then check if there is the target in the build file
	bf, err := ParseBuildFile(targetFilePath, w.rootPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("parse build file: %w", err)
	}

	// if build file does not exist then create it
	if errors.Is(err, os.ErrNotExist) {
		// create a build file
		f, err := os.Create(targetFilePath)
		if err != nil {
			return fmt.Errorf("create build file: %w", err)
		}
		bf, err = CreateBuildFile(targetFilePath, getPackageName(target), f)
		if err != nil {
			return fmt.Errorf("create build file: %w", err)
		}
		f.Close()
	}

	targetName := getTargetName(target)
	var hasTarget bool
	if bf != nil {
		hasTarget = bf.HasTarget(targetName)
	}
	// if there is then throw an error
	if hasTarget {
		return ErrTargetExists
	}

	// add target to build file/
	if err := bf.CreateTarget(targetName,
		strings.NewReader(GetImportedCommand(args))); err != nil {
		return fmt.Errorf("create target: %w", err)
	}

	// run the target
	if err := w.RunTarget(ctx, target); err != nil {
		return fmt.Errorf("run target: %w", err)
	}
	return nil
}
