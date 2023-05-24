package workspace

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aakarim/mmake/internal/makefile"
)

type Label string

const RootLabel = "//"

type BuildFile struct {
	// path to the build file
	Path string
	// label of the build file
	Label Label
	// list of targets in the build file
	Targets []string
	// the description of the build file (if any)
	Description string
}

func (b *BuildFile) HasTarget(name string) bool {
	for _, t := range b.Targets {
		if t == name {
			return true
		}
	}
	return false
}

func CreateBuildFile(path, label string, w io.Writer) (*BuildFile, error) {
	return &BuildFile{
		Path:  path,
		Label: Label(label),
	}, nil
}

func FileIsBuildFile(path string) bool {
	if f := filepath.Base(path); strings.HasSuffix(f, "makefile") ||
		strings.HasSuffix(f, "Makefile") {
		return true
	}
	return false
}

func ParseBuildFile(path string, rootDir string) (*BuildFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	str, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var desc string
	// if the first character is a #, then this is a description
	if len(str) > 0 && str[0] == '#' {
		// find the first newline
		for i, c := range str {
			if c == '\n' {
				desc = string(str[:i])
				break
			}
		}
	}

	mf, err := makefile.ParseMakefile(strings.NewReader(string(str)))
	if err != nil {
		return nil, err
	}
	label, err := GetPackageFromFile(path, rootDir)
	if err != nil {
		return nil, err
	}

	// parse as makefile
	return &BuildFile{
		Path:  path,
		Label: Label(label),
		Targets: func() []string {
			var targets []string
			targets = append(targets, mf.Targets...)
			return targets
		}(),
		Description: desc,
	}, nil
}

// CreateTarget creates a new target in the build file
// using the reader as the content of the target
// if the target already exists, then it will throw an error.
// opens a file handle to the build file.
func (bf *BuildFile) CreateTarget(name string, targetBody io.Reader) error {
	f, err := os.OpenFile(bf.Path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open build file: %w", err)
	}

	defer f.Close()

	b, err := io.ReadAll(targetBody)
	if err != nil {
		return fmt.Errorf("read target body: %w", err)
	}

	builder := strings.Builder{}
	builder.Write([]byte("\n\n"))
	builder.Write([]byte(name))
	// the target body needs a tab
	builder.Write([]byte(":\n\t"))
	builder.Write(b)
	builder.Write([]byte("\n"))

	if _, err := f.Write([]byte(builder.String())); err != nil {
		return fmt.Errorf("write to build file: %w", err)
	}
	return nil
}
