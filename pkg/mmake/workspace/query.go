package workspace

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	if prefix[:2] == RootLabel {
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
	// if does not start with // then it's not valid
	if len(prefix) < 2 || prefix[:2] != RootLabel {
		return "", &ErrInvalidQuery{query: prefix, message: "prefix must start with //"}
	}

	// if there is a ':' in the prefix, then complete to targets only
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

	// otherwise complete to files and append the targets of the current label
	completions, err := q.genCompFiles(ctx, prefix)
	if err != nil {
		return "", err
	}

	bf := q.GetFileByLabel(Label(prefix))
	if bf != nil {
		label, targets, err := q.genCompTargets(ctx, prefix+":")
		if err != nil {
			return "", err
		}
		for _, v := range targets {
			completions = append(completions, fmt.Sprintf("%s:%s", string(label), v))
		}
	}

	var outputStr string
	for _, completion := range completions {
		outputStr += string(completion) + "\n"
	}
	return outputStr, nil
}

func (q *Query) GetFileByLabel(label Label) *BuildFile {
	for _, v := range q.files {
		if v.Label == label {
			return v
		}
	}

	return nil
}

func (q *Query) genCompTargets(ctx context.Context, prefix string) (Label, []string, error) {
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
		return RootLabel, q.files[0].Targets, nil
	}

	spl := strings.Split(prefix, ":")
	if len(spl) != 2 {
		return "", nil, &ErrInvalidQuery{query: prefix, message: "prefix must contain a ':'"}
	}

	// get the file that matches the prefix
	var file *BuildFile
	for _, f := range q.files {
		if string(f.Label) == spl[0] {
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

func (q *Query) genCompFiles(ctx context.Context, prefix string) ([]string, error) {
	if prefix == "" {
		return nil, &ErrInvalidQuery{query: prefix, message: "prefix required"}
	}
	if len(prefix) < 2 {
		return nil, &ErrInvalidQuery{query: prefix, message: "prefix must be at least 2 characters"}
	}
	// if the prefix is already a label that is useful information
	var isLabel bool
	bf := q.GetFileByLabel(Label(prefix))
	if bf != nil {
		isLabel = true
	}

	// strip //
	if prefix[:2] == RootLabel {
		prefix = prefix[2:]
	}
	// get directory of the prefix and compare to the directory of the file
	// if they match, then add the file to the list
	prefixPath := path.Join(q.ws.rootPath, prefix)

	// get the labels and directories that are potential completions for the prefix
	// This could be:
	// 1. The prefix itself as a Label
	// 2. The prefix as a directory (with a trailing '/')
	// 3. The prefix + the intervening directories until the next label e.g. prefix: "//" - ["//", "//pkg/foo", "//pkg/bar"] -> ["//", "//pkg/"]

	// copy & sort the files by descending path length
	files := make([]*BuildFile, len(q.files))
	copy(files, q.files)

	sort.Slice(files, func(i, j int) bool {
		return len(files[i].Path) > len(files[j].Path)
	})

	interveningDirs := map[string]bool{}
	var thisLevelCompletions []string
	for _, f := range files {
		currentDir := path.Dir(f.Path)
		relDirPath := strings.TrimPrefix(currentDir, q.ws.rootPath)

		// invariant: prefixPath is a prefix of currentDir
		if !strings.HasPrefix(currentDir, prefixPath) {
			continue
		}

		// if the file will match up the prefix path, then add it to the list
		var hasThisLevelCompletions bool
		for _, v := range thisLevelCompletions {
			enclosingDir := path.Dir(v)
			if path.Dir(relDirPath) == enclosingDir {
				hasThisLevelCompletions = true
				break
			}
		}
		if !hasThisLevelCompletions {
			// if the same level then add
			if path.Dir(currentDir) == path.Dir(prefixPath) {
				thisLevelCompletions = []string{}
				thisLevelCompletions = append(thisLevelCompletions, string(f.Label))
			}
			if !isLabel {
				// if the labels are not in the same directory then assume
				// that this is lower level and reset the list
				if len(thisLevelCompletions) > 0 {
					if path.Dir(currentDir) != path.Dir(thisLevelCompletions[0]) {
						thisLevelCompletions = []string{}
					}
				}
				thisLevelCompletions = append(thisLevelCompletions, string(f.Label))
			}
		} else {
			thisLevelCompletions = append(thisLevelCompletions, string(f.Label))
		}

		if !isLabel || path.Dir(f.Path) == prefixPath {
			continue
		}

		if interveningDirs[relDirPath] {
			delete(interveningDirs, relDirPath)
		}

		// remove the last directory from the path
		// e.g. /pkg/foo/bar -> /pkg/foo
		relDirPath = path.Dir(relDirPath)

		// if the path is shorter than the prefix, then skip it
		if len(relDirPath) < len(prefix) {
			continue
		}
		// if the prefix is a directory, then add it to the list
		if relDirPath != "/" {
			interveningDirs[relDirPath] = true
		}
	}

	completions := make([]string, 0, len(thisLevelCompletions)+len(interveningDirs))
	// add the primary completion to the list of completions
	if len(thisLevelCompletions) != 0 {
		completions = append(completions, thisLevelCompletions...)
	}

	// add the label directories to the list of completions
	for labelDir := range interveningDirs {
		completions = append(completions, "/"+labelDir+"/")
	}

	// sort by length and then lexically
	sort.Slice(completions, func(i, j int) bool {
		if len(completions[i]) == len(completions[j]) {
			return completions[i] < completions[j]
		}
		return len(completions[i]) < len(completions[j])
	})

	// remove duplicates
	seen := map[string]bool{}
	for i := 0; i < len(completions); i++ {
		if seen[completions[i]] {
			completions = append(completions[:i], completions[i+1:]...)
			i--
		}
		seen[completions[i]] = true
	}

	return completions, nil
}

func GetPackageFromFile(filePath string, rootPath string) (string, error) {
	// get the directory of the file
	dir := filepath.Dir(filePath)
	if dir == rootPath {
		return RootLabel, nil
	}
	dir = dir[len(rootPath):]

	// strip the leading '/'
	dir = dir[1:]

	// prepend '//'
	return RootLabel + dir, nil
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
// and re-parsing all of the Makefiles.
// Depth is the depth of the directory tree to scan. If depth is 0, then
// the entire tree is scanned.
func (q *Query) Update(ctx context.Context, depth int) error {
	// clear the list of files
	q.files = nil
	// walk the workspace directory and find all the Makefiles
	return filepath.WalkDir(q.ws.rootPath, func(path string, d os.DirEntry, err error) error {
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
}
