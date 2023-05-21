package workspace

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var ErrInvalidQuery = fmt.Errorf("invalid query")

// QueryByPrefix returns a list of workspace-aware files that match the given prefix
func (q *Query) QueryFilesByPrefix(ctx context.Context, prefix string) ([]*BuildFile, error) {
	if prefix == "" {
		return nil, ErrInvalidQuery
	}
	if len(prefix) < 2 {
		return nil, ErrInvalidQuery
	}
	// strip //
	if prefix[:2] == "//" {
		prefix = prefix[2:]
	}
	// get directory of the prefix and compare to the directory of the file
	// if they match, then add the file to the list
	prefixPath := path.Join(q.ws.rootPath, prefix)

	// if the prefix is the root directory, then return the root Makefile
	if prefixPath == q.ws.rootPath {
		return []*BuildFile{q.files[0]}, nil
	}

	var files []*BuildFile
	for _, f := range q.files {
		if len(f.Path) < len(prefixPath) {
			continue
		}
		if f.Path[:len(prefixPath)] != prefixPath {
			continue
		}

		files = append(files, f)
	}

	return files, nil
}

func (q *Query) GetPackageFromFile(filePath string) (string, error) {
	// get the directory of the file
	dir := filepath.Dir(filePath)
	if dir == q.ws.rootPath {
		return "//", nil
	}
	dir = dir[len(q.ws.rootPath):]

	// strip the leading '/'
	dir = dir[1:]

	// prepend '//'
	return "//" + dir, nil
}

// QueryTargetsByFile returns a list of targets that are contained within the given file
func (w *Query) QueryTargetsByFile(ctx context.Context, filePath string) ([]string, error) {
	return nil, nil
}

type Query struct {
	ws *Workspace
	// the list of workspace-aware files in the workspace
	files []*BuildFile
}

func NewQuery(ws *Workspace) *Query {
	return &Query{ws: ws}
}

// Update updates the workspace by re-scanning the workspace directory
func (q *Query) Update(ctx context.Context) error {
	// clear the list of files
	q.files = nil
	// walk the workspace directory and find all the Makefiles
	filepath.WalkDir(q.ws.rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && FileIsBuildFile(d.Name()) {
			// parse
			f, err := ParseBuildFile(path)
			if err != nil {
				return fmt.Errorf("failed to parse build file %s: %w", path, err)
			}
			q.files = append(q.files, f)
		}

		if d.IsDir() && d.Name() == BuildDir {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		return nil
	})

	return nil
}
