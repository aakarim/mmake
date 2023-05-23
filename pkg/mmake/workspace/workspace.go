package workspace

import (
	"context"
	"os"
	"path"
)

type Workspace struct {
	rootPath   string
	ignoreDirs []string
}

func New(rootPath string) *Workspace {
	return &Workspace{rootPath: rootPath, ignoreDirs: []string{
		".git",
		"build-out",
		"vendor",
		"node_modules",
		"__pycache__",
		"__snapshots__",
		"__tests__",
		"__mocks__",
		"__fixtures__",
	}}
}

func (w *Workspace) Init(ctx context.Context) error {
	// create the build-out directory if it doesn't exist
	if err := os.MkdirAll(path.Join(w.rootPath, "build-out"), 0755); err != nil {
		return err
	}

	return nil
}
