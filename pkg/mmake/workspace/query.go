package workspace

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ErrInvalidQuery struct {
	query   string
	message string
}

func (e *ErrInvalidQuery) Error() string {
	return fmt.Sprintf("invalid query: %s", e.message)
}

// QueryByPrefix returns a list of workspace-aware files that match the given prefix
func (q *Query) QueryFilesByPrefix(ctx context.Context, prefix string) ([]*BuildFile, error) {
	if prefix == "" {
		return nil, &ErrInvalidQuery{query: prefix, message: "prefix required"}
	}
	if len(prefix) < 2 {
		return nil, &ErrInvalidQuery{query: prefix, message: "prefix must be at least 2 characters"}
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

func (q *Query) printFiles(files []*BuildFile) error {
	// print the files as columns with file and description
	for _, file := range files {
		packageName, err := GetPackageFromFile(file.Path, q.ws.rootPath)
		if err != nil {
			return err
		}
		fmt.Printf("%s\t%s\n", packageName, file.Description)
	}
	return nil
}

// GenComp completes the given prefix to a list of files and targets that match the prefix
// if there is a ':' in the input, it will complete to targets, otherwise it will complete to files.
// TODO: move this to the 'completion' package.
func (q *Query) GenComp(ctx context.Context, prefix string) (string, error) {
	if prefix == "" {
		return "", &ErrInvalidQuery{query: prefix, message: "prefix required"}
	}
	// if does not start with // then it's not valid
	if len(prefix) < 2 && prefix[:2] != "//" {
		return "", &ErrInvalidQuery{query: prefix, message: "prefix must be at least 2 characters and start with //"}
	}

	// if there is a ':' in the prefix, then complete to targets
	if strings.Contains(prefix, ":") {
		label, targets, err := q.genCompTargets(ctx, prefix)
		if err != nil {
			return "", err
		}
		var outputStr string
		for _, target := range targets {
			outputStr += fmt.Sprintf("%s:%s\n", label, target)
		}
		return outputStr, nil
	}

	// otherwise complete to files
	files, err := q.genCompFiles(ctx, prefix)
	if err != nil {
		return "", err
	}
	var outputStr string
	for _, file := range files {
		p, err := GetPackageFromFile(file.Path, q.ws.rootPath)
		if err != nil {
			return "", err
		}
		outputStr += p + "\n"
	}
	return outputStr, nil
}

func (q *Query) genCompTargets(ctx context.Context, prefix string) (string, []string, error) {
	if prefix == "" {
		return "", nil, &ErrInvalidQuery{query: prefix, message: "prefix required"}
	}
	if len(prefix) < 3 {
		return "", nil, &ErrInvalidQuery{query: prefix, message: "prefix must be at least 3 characters"}
	}

	// ensure there is a ':' in the prefix
	if !strings.Contains(prefix, ":") {
		return "", nil, &ErrInvalidQuery{query: prefix, message: "prefix must contain a ':'"}
	}

	// get directory of the prefix and compare to the directory of the file
	// if they match, then add the file to the list
	prefixPath := path.Join(q.ws.rootPath, prefix)

	// if the prefix is the root directory, then return the targets in the root Makefile
	if prefixPath == q.ws.rootPath {
		return "//", q.files[0].Targets, nil
	}

	spl := strings.Split(prefix, ":")
	if len(spl) != 2 {
		return "", nil, &ErrInvalidQuery{query: prefix, message: "prefix must contain a ':'"}
	}

	// get the file that matches the prefix
	var file *BuildFile
	for _, f := range q.files {
		if f.Label == spl[0] {
			file = f
			break
		}
	}

	if file == nil {
		return "", nil, &ErrInvalidQuery{query: prefix, message: "no file found for prefix"}
	}
	label := file.Label

	// get the targets that match the prefix
	var targets []string
	for _, t := range file.Targets {
		if strings.HasPrefix(t, spl[1]) {
			targets = append(targets, t)
		}
	}

	return label, targets, nil
}

func (q *Query) genCompFiles(ctx context.Context, prefix string) ([]*BuildFile, error) {
	if prefix == "" {
		return nil, &ErrInvalidQuery{query: prefix, message: "prefix required"}
	}
	if len(prefix) < 2 {
		return nil, &ErrInvalidQuery{query: prefix, message: "prefix must be at least 2 characters"}
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

func GetPackageFromFile(filePath string, rootPath string) (string, error) {
	// get the directory of the file
	dir := filepath.Dir(filePath)
	if dir == rootPath {
		return "//", nil
	}
	dir = dir[len(rootPath):]

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
			f, err := ParseBuildFile(path, q.ws.rootPath)
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
