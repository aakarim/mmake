package workspace

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aakarim/mmake/pkg/mmake/makefile"
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
