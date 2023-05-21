package completion

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/aakarim/mmake/pkg/mmake/workspace"
)

type Completion struct {
	workspace *workspace.Workspace
}

func New(workspace *workspace.Workspace) *Completion {
	return &Completion{
		workspace: workspace,
	}
}

// GetCompletionScript returns the completion script for the current shell
func GetCompletionScript(workspacePath string) string {
	template := template.Must(template.New("completion").Parse(completionScript))
	bufStr := bytes.NewBufferString("")
	if err := template.Execute(bufStr, templateVars{
		WorkspaceDir: filepath.Dir(workspacePath),
	}); err != nil {
		panic(err)
	}

	return bufStr.String()
}
