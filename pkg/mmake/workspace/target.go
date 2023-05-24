package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/aakarim/mmake/internal/makefile"
)

func (w *Workspace) RunTarget(ctx context.Context, target string) error {
	targetFilePath, err := w.getBuildFile(ctx, target)
	if err != nil {
		return err
	}
	targetName := getTargetName(target)
	var args []string
	if targetFilePath != "" {
		args = append(args, "-f", targetFilePath)
		if targetName != "" {
			args = append(args, targetName)
		}
	}

	cmd := exec.CommandContext(ctx, "make", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	envVars, err := w.buildEnv(targetFilePath)
	if err != nil {
		return err
	}
	cmd.Env = append(os.Environ(), envVars...)
	if err := cmd.Run(); err != nil {
		return &ErrCommand{err}
	}
	return nil
}

func getTargetName(target string) string {
	// strip out leading '//'
	target = target[2:]
	// split on ':'
	if strings.Contains(target, ":") {
		return strings.Split(target, ":")[1]
	}
	return target
}

func getPackageName(target string) string {
	// strip out leading '//'
	target = target[2:]
	// split on ':'
	return strings.Split(target, ":")[0]
}

func getRelPathFromTarget(target string) string {
	// strip out leading '//'
	target = target[2:]
	// split on ':'
	return strings.Split(target, ":")[0]
}

func (w *Workspace) getBuildFile(ctx context.Context, target string) (string, error) {
	target = getRelPathFromTarget(target)

	targetFilePath := ""
	// check if Makefile exists in dir
	if err := filepath.WalkDir(filepath.Join(w.rootPath, target), func(path string, d os.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err != nil {
				return err
			}
			if !d.IsDir() && FileIsBuildFile(path) {
				targetFilePath = path
				return filepath.SkipDir
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	if targetFilePath == "" {
		return "", ErrNoMakefileFound
	}
	return filepath.Abs(targetFilePath)
}

func (w *Workspace) GetInfo(ctx context.Context, target string) (string, error) {
	bf, err := w.getBuildFile(ctx, target)
	if err != nil {
		return "", fmt.Errorf("get build file: %w", err)
	}

	// load build file and get info
	f, err := os.Open(bf)
	if err != nil {
		return "", fmt.Errorf("open build file: %w", err)
	}
	defer f.Close()

	targetName := getTargetName(target)

	t := makefile.GetTarget(targetName, f)
	if t == nil {
		return "", fmt.Errorf("target not found: %s", targetName)
	}
	if t.Body != "" {
		// if the first line is a comment, show the comment until the first blank line
		if strings.HasPrefix(t.Body, "#") {
			lines := strings.Split(t.Body, "\n")
			for i, l := range lines {
				if l == "" {
					t.Body = strings.Join(lines[:i], "\n")
					break
				}
			}
		}
		return t.Body, nil
	}
	return "no target body", nil
}

func (w *Workspace) buildEnv(targetFilePath string) ([]string, error) {
	rel, err := filepath.Rel(w.rootPath, filepath.Dir(targetFilePath))
	if err != nil {
		return nil, err
	}
	buildTargetDir := path.Join(w.rootPath, BuildDir, rel)
	if err := os.MkdirAll(buildTargetDir, 0755); err != nil {
		return nil, err
	}

	return []string{
		"MM_ROOT=" + filepath.Join(w.rootPath, "WORKSPACE.mmake"),
		"MM_PATH=" + filepath.Join(w.rootPath, rel),
		"MM_OUT_ROOT=" + path.Join(w.rootPath, BuildDir),
		"MM_OUT_PATH=" + buildTargetDir,
	}, nil
}

func (w *Workspace) Clean(ctx context.Context, label string) error {
	if label == "" {
		return errors.New("you must specify a label to clean")
	}

	// strip out leading '//'
	label = label[2:]

	buildTargetDir := path.Join(w.rootPath, BuildDir, label)
	if err := os.RemoveAll(buildTargetDir); err != nil {
		return err
	}
	return nil
}
