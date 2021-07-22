package tools

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"mvdan.cc/sh/v3/syntax"

	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/field"
	"arhat.dev/dukkha/pkg/renderer"
	"arhat.dev/dukkha/pkg/sliceutils"
)

type Action struct {
	field.BaseField

	// Name of this action, optional
	Name string `yaml:"name"`

	// Task reference of this action
	//
	// Task, Cmd, Shell, OtherShell are mutually exclusive
	Task string `yaml:"task"`

	// Shell using embedded shell
	//
	// Task, Cmd, Shell, OtherShell are mutually exclusive
	EmbeddedShell string `yaml:"shell"`

	// Shell script for this action
	//
	// Task, Cmd, Shell, OtherShell are mutually exclusive
	OtherShell map[string]string `dukkha:"other"`

	// Cmd execution, not in any shell
	//
	// Task, Cmd, Shell, OtherShell are mutually exclusive
	Cmd []string `yaml:"cmd"`

	// Chdir change working directory before executing command
	// this option only applies to Cmd, Shell, OtherShell action
	Chdir string `yaml:"chdir"`

	// ContuineOnError ignores error occurred in this action and continue
	// following actions in list (if any)
	ContinueOnError bool `yaml:"continue_on_error"`

	mu sync.Mutex
}

func (act *Action) DoAfterFieldResolved(mCtx dukkha.TaskExecContext, do func(h *Action) error) error {
	act.mu.Lock()
	defer act.mu.Unlock()

	err := act.ResolveFields(mCtx, -1, "")
	if err != nil {
		return fmt.Errorf("failed to resolve fields: %w", err)
	}

	return do(act)
}

func (act *Action) GenSpecs(
	ctx dukkha.TaskExecContext, index int,
) (dukkha.RunTaskOrRunCmd, error) {
	hookID := "#" + strconv.FormatInt(int64(index), 10)
	if len(act.Name) != 0 {
		hookID = fmt.Sprintf("%s (%s)", act.Name, hookID)
	}

	switch {
	case len(act.Task) != 0:
		return act.genTaskActionSpecs(ctx, hookID)
	case len(act.Cmd) != 0:
		return act.genCmdActionSpecs(ctx, hookID)
	case len(act.EmbeddedShell) != 0:
		return act.genEmbeddedShellActionSpecs(ctx, hookID)
	default:
		return act.genShellActionSpecs(ctx, hookID)
	}
}

func (act *Action) genTaskActionSpecs(
	ctx dukkha.TaskExecContext, hookID string,
) (dukkha.RunTaskOrRunCmd, error) {
	ref, err := dukkha.ParseTaskReference(act.Task, ctx.CurrentTool().Name)
	if err != nil {
		return nil, fmt.Errorf("%q: invalid task reference %q: %w", hookID, act.Task, err)
	}

	if len(ref.MatrixFilter) != 0 {
		ctx.SetMatrixFilter(ref.MatrixFilter)
	}

	tool, ok := ctx.GetTool(ref.ToolKey())
	if !ok {
		return nil, fmt.Errorf("%q: referenced tool %q not found", hookID, ref.ToolKey())
	}

	tsk, ok := tool.GetTask(ref.TaskKey())
	if !ok {
		return nil, fmt.Errorf("%q: referenced task %q not found", hookID, ref.TaskKey())
	}

	return &TaskExecRequest{
		Context:     ctx,
		Tool:        tool,
		Task:        tsk,
		IgnoreError: act.ContinueOnError,
	}, nil
}

func (act *Action) genCmdActionSpecs(
	ctx dukkha.TaskExecContext, hookID string,
) (dukkha.RunTaskOrRunCmd, error) {
	_ = hookID
	return []dukkha.TaskExecSpec{
		{
			Env:         sliceutils.FormatStringMap(ctx.Env(), "=", false),
			Command:     sliceutils.NewStrings(act.Cmd),
			Chdir:       act.Chdir,
			IgnoreError: act.ContinueOnError,
		},
	}, nil
}

func (act *Action) genEmbeddedShellActionSpecs(
	ctx dukkha.TaskExecContext, hookID string,
) (dukkha.RunTaskOrRunCmd, error) {

	workingDir := act.Chdir
	script := act.EmbeddedShell

	return []dukkha.TaskExecSpec{{
		AlterExecFunc: func(
			replace map[string][]byte,
			stdin io.Reader,
			stdout, stderr io.Writer,
		) (dukkha.RunTaskOrRunCmd, error) {
			environ, _ := renderer.CreateEnvForEmbeddedShell(ctx)
			runner, err := renderer.CreateEmbeddedShellRunner(
				workingDir, environ, stdin, stdout, stderr,
			)
			if err != nil {
				return nil, fmt.Errorf("%q: failed to create embedded shell: %w", hookID, err)
			}

			parser := syntax.NewParser(syntax.Variant(syntax.LangBash))

			err = renderer.RunShellScriptInEmbeddedShell(ctx, runner, parser, script)
			if err != nil {
				return nil, fmt.Errorf("%q: failed to run command in embedded shell: %w", hookID, err)
			}

			return nil, err
		},
	}}, nil
}

func (act *Action) genShellActionSpecs(
	ctx dukkha.TaskExecContext, hookID string,
) (dukkha.RunTaskOrRunCmd, error) {
	// check other shell
	switch {
	case len(act.OtherShell) > 1:
		return nil, fmt.Errorf(
			"%q: unexpected multiple shell entries in one spec",
			hookID,
		)
	case len(act.OtherShell) == 1:
		// ok
	default:
		// no hook to run
		return nil, nil
	}

	var (
		shell  string
		script string
	)

	for k, v := range act.OtherShell {
		script = v

		switch {
		case strings.HasPrefix(k, "shell:"):
			shell = strings.SplitN(k, ":", 2)[1]
		default:
			return nil, fmt.Errorf("%q: unknown action: %q", hookID, k)
		}
	}

	return []dukkha.TaskExecSpec{
		{
			Env:         sliceutils.FormatStringMap(ctx.Env(), "=", false),
			Command:     []string{script},
			Chdir:       act.Chdir,
			UseShell:    true,
			ShellName:   shell,
			IgnoreError: act.ContinueOnError,
		},
	}, nil
}