package makefile

import (
	"bufio"
	"fmt"
	"go/scanner"
	"io"
	"strings"
)

type Makefile struct {
	Targets []string
}

var internalTargets = []string{
	".PHONY",
	".SUFFIXES",
	".DEFAULT",
	".PRECIOUS",
	".INTERMEDIATE",
	".SECONDARY",
	".SECONDEXPANSION",
	".DELETE_ON_ERROR",
	".IGNORE",
	".LOW_RESOLUTION_TIME",
	".SILENT",
	".EXPORT_ALL_VARIABLES",
	".NOTPARALLEL",
	".ONESHELL",
	".POSIX",
	".SHELLFLAGS",
	".FEATURES",
	".MAKE",
	".FEATURES",
	".VARIABLES",
	".MAKEFILE_LIST",
	".MAKEFLAGS",
	".TARGETS",
	".THIS_FILE",
	".DEFAULT_GOAL",
	".RECIPEPREFIX",
	".SHELL",
	".LIBPATTERNS",
}

func ParseMakefile(file io.Reader) (*Makefile, error) {
	mf := Makefile{}
	// get the targets from the makefile
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		// targets follow the format [target]: [dependencies]
		// we only care about the target
		scanned := scan.Text()
		if len(scanned) == 0 {
			continue
		}
		if scanned[0] == '#' {
			continue
		}
		if scanned[0] == '.' {
			continue
		}
		if scanned[0] == '\t' {
			continue
		}
		if scanned[0] == ' ' {
			continue
		}
		if scanned[0] == '\n' {
			continue
		}
		if scanned[0] == '\r' {
			continue
		}

		// check if this is an internal target
		for _, internalTarget := range internalTargets {
			if scanned == internalTarget {
				continue
			}
		}

		// check if this is a variable
		if scanned[0] == '$' {
			continue
		}

		// check if this is a function
		if scanned[0] == '%' {
			continue
		}

		// if the line contains a string up until ': ' or ':\n' then it is a target
		if isTarget(scanned) {
			target := strings.Split(scanned, ":")[0]
			mf.Targets = append(mf.Targets, target)
		}
	}
	if err := scan.Err(); err != nil {
		if err, ok := err.(*scanner.Error); ok {
			return nil, fmt.Errorf("%s:%d:%d: %s", err.Pos.Filename, err.Pos.Line, err.Pos.Column, err.Msg)
		}
		return nil, fmt.Errorf("error scanning makefile: %w", err)
	}

	return &mf, nil
}

func isTarget(str string) bool {
	// if the line contains a string up until ': ' or ':\n' then it is a target
	if (strings.Contains(str, ": ") || strings.Contains(str, ":")) && !strings.Contains(str, " :") {
		return true
	}
	return false
}

func GetTarget(name string, f io.Reader) *Target {
	scan := bufio.NewScanner(f)
	var target *Target
	for scan.Scan() {
		scanned := scan.Text()
		if len(scanned) == 0 {
			continue
		}
		if target != nil {
			// if we've reached the end of the target, return it
			if scanned[0] != '\t' {
				return target
			}

			target.Body += scanned + "\n"
			continue
		}

		if scanned[0] == '#' {
			continue
		}
		if scanned[0] == '.' {
			continue
		}
		if scanned[0] == '\t' {
			continue
		}
		if scanned[0] == ' ' {
			continue
		}
		if scanned[0] == '\n' {
			continue
		}
		if scanned[0] == '\r' {
			continue
		}

		// check if this is an internal target
		for _, internalTarget := range internalTargets {
			if scanned == internalTarget {
				continue
			}
		}

		// check if this is a variable
		if scanned[0] == '$' {
			continue
		}

		// check if this is a function
		if scanned[0] == '%' {
			continue
		}

		// if the line contains a string up until ': ' or ':\n' then it is a target
		if isTarget(scanned) {
			targetName := strings.Split(scanned, ":")[0]
			if targetName == name {
				target = &Target{
					Name: name,
				}
			}
		}
	}

	return target
}

type Target struct {
	Name string
	Body string
}
