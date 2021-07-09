package docker

import (
	"strings"

	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/sliceutils"
	"arhat.dev/dukkha/pkg/tools/buildah"
)

const TaskKindLogin = "login"

func init() {
	dukkha.RegisterTask(
		ToolKind, TaskKindLogin,
		func(toolName string) dukkha.Task {
			t := &TaskLogin{}
			t.InitBaseTask(ToolKind, dukkha.ToolName(toolName), TaskKindLogin)
			return t
		},
	)
}

type TaskLogin buildah.TaskLogin

func (c *TaskLogin) GetExecSpecs(
	rc dukkha.RenderingContext,
	useShell bool,
	shellName string,
	dockerCmd []string,
) ([]dukkha.TaskExecSpec, error) {
	loginCmd := sliceutils.NewStrings(
		dockerCmd, "login",
		"--username", c.Username,
		"--password-stdin",
	)

	password := c.Password + "\n"
	return []dukkha.TaskExecSpec{{
		Stdin:       strings.NewReader(password),
		Command:     append(loginCmd, c.Registry),
		IgnoreError: false,
		UseShell:    useShell,
		ShellName:   shellName,
	}}, nil
}
