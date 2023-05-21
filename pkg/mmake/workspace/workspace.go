package workspace

import (
	"context"
	"os"
	"path"
)

type Workspace struct {
	rootPath string
}

func New(rootPath string) *Workspace {
	return &Workspace{rootPath: rootPath}
}

func (w *Workspace) Init(ctx context.Context) error {
	// create the build-out directory if it doesn't exist
	if err := os.MkdirAll(path.Join(w.rootPath, "build-out"), 0755); err != nil {
		return err
	}

	return nil
}
