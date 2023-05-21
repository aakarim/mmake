package workspace

import (
	"context"
	"os"
	"path"
	"path/filepath"
)

// FindWorkspaceFile finds the WORKSPACE.mmake relative to the current directory
// if a path is specified, it will search relative to that path.
// If the path is a WORKSPACE.mmake file, it will return that path.
func FindWorkspaceFile(ctx context.Context, inputPath string) (string, error) {
	// search for the WORKSPACE.mmake file in current and all parent directories
	rootPath := "."
	if inputPath != "" {
		if p := filepath.Base(inputPath); p == "WORKSPACE.mmake" {
			return inputPath, nil
		}
		rootPath = inputPath
	}

	var workspacePath string
	for workspacePath == "" {
		err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err != nil {
					return err
				}

				if !d.IsDir() && d.Name() == "WORKSPACE.mmake" {
					workspacePath, err = filepath.Abs(path)
					if err != nil {
						return err
					}
					return filepath.SkipDir
				}
				return nil
			}
		})
		if err != nil {
			return "", err
		}

		if rootPath == "/" {
			return "", ErrNoWorkspaceFound
		}

		// go to parent
		rootPath = path.Clean(path.Join("../", rootPath))
	}
	return workspacePath, nil
}
