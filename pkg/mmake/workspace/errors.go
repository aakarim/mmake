package workspace

import "errors"

var ErrNoWorkspaceFound = errors.New("no WORKSPACE.mmake file found")

var ErrNoMakefileFound = errors.New("no Makefile found")

type ErrCommand struct {
	Err error
}

func (e *ErrCommand) Error() string {
	return e.Err.Error()
}

func (e *ErrCommand) Unwrap() error {
	return e.Err
}
