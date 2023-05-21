package completion

import (
	_ "embed"
)

//go:embed completion.bash.tmpl
var completionScript string

type templateVars struct {
	WorkspaceDir string
}
