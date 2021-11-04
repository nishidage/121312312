package kubectl

import (
	"arhat.dev/rs"

	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/tools"
)

const ToolKind = "kubectl"

func init() {
	dukkha.RegisterTool(ToolKind, func() dukkha.Tool {
		return &Tool{}
	})
}

type Tool struct {
	rs.BaseField `yaml:"-"`

	tools.BaseTool `yaml:",inline"`
}

func (t *Tool) Init(kind dukkha.ToolKind, cacheDir string) error {
	return t.InitBaseTool(kind, "kubectl", cacheDir, t)
}
