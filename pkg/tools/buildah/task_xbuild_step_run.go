package buildah

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"arhat.dev/pkg/md5helper"
	"arhat.dev/rs"

	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/sliceutils"
)

type stepRun struct {
	rs.BaseField

	// Script
	Script string `yaml:"script"`

	// ScriptArgs for the script
	ScriptArgs []string `yaml:"script_args"`

	Cmd []string `yaml:"cmd"`

	ExtraArgs []string `yaml:"extra_args"`
}

func (s *stepRun) genSpec(
	rc dukkha.TaskExecContext,
	options dukkha.TaskMatrixExecOptions,
	record bool,
) ([]dukkha.TaskExecSpec, error) {
	runCmd := sliceutils.NewStrings(options.ToolCmd(), "run")
	if record {
		runCmd = append(runCmd, "--add-history")
	}
	runCmd = append(runCmd, s.ExtraArgs...)
	runCmd = append(runCmd, replace_XBUILD_CURRENT_CONTAINER_ID, "--")

	var steps []dukkha.TaskExecSpec

	switch {
	case len(s.Cmd) != 0:
		steps = append(steps, dukkha.TaskExecSpec{
			IgnoreError: false,
			Command:     append(runCmd, s.Cmd...),
			UseShell:    options.UseShell(),
			ShellName:   options.ShellName(),
		})
	case len(s.Script) != 0:
		// copy this script to container
		const (
			replace_XBUILD_RUN_SCRIPT_SRC_PATH          = "<XBUILD_RUN_SCRIPT_SRC_PATH>"
			replace_XBUILD_RUN_SCRIPT_SRC_REDACTED_PATH = "<XBUILD_RUN_SCRIPT_SRC_REDACTED_PATH>"
		)

		script := s.Script
		steps = append(steps,
			// write script to local cache
			dukkha.TaskExecSpec{
				StdoutAsReplace:          replace_XBUILD_RUN_SCRIPT_SRC_PATH,
				FixStdoutValueForReplace: bytes.TrimSpace,

				ShowStdout: true,
				AlterExecFunc: func(
					replace dukkha.ReplaceEntries,
					stdin io.Reader,
					stdout, stderr io.Writer,
				) (dukkha.RunTaskOrRunCmd, error) {
					srcFile := filepath.Join(
						rc.CacheDir(),
						"buildah", "xbuild",
						"run-script-"+hex.EncodeToString(md5helper.Sum([]byte(script))),
					)
					err := os.MkdirAll(filepath.Dir(srcFile), 0755)
					if err != nil {
						return nil, fmt.Errorf("failed to ensure script cache dir: %w", err)
					}

					_, err = stdout.Write([]byte(srcFile))
					if err != nil {
						return nil, fmt.Errorf("failed to create script cache: %q", srcFile)
					}

					return nil, os.WriteFile(srcFile, []byte(script), 0644)
				},
			},
			// write redacted file
			dukkha.TaskExecSpec{
				StdoutAsReplace:          replace_XBUILD_RUN_SCRIPT_SRC_REDACTED_PATH,
				FixStdoutValueForReplace: bytes.TrimSpace,

				AlterExecFunc: func(
					replace dukkha.ReplaceEntries,
					stdin io.Reader,
					stdout, stderr io.Writer,
				) (dukkha.RunTaskOrRunCmd, error) {
					v, ok := replace[replace_XBUILD_RUN_SCRIPT_SRC_PATH]
					if !ok {
						return nil, fmt.Errorf("unexpected script path not found")
					}

					srcFile := string(v.Data)
					redactedSrcFile := srcFile + "-redacted"
					_, err := stdout.Write([]byte(redactedSrcFile))
					if err != nil {
						return nil, fmt.Errorf("failed to write redacted file path: %w", err)
					}

					return nil, os.WriteFile(redactedSrcFile, []byte(""), 0644)
				},
			},
			// copy script to container
			dukkha.TaskExecSpec{
				Command: sliceutils.NewStrings(
					options.ToolCmd(), "copy", "--chmod", "0755",
					replace_XBUILD_CURRENT_CONTAINER_ID,
					replace_XBUILD_RUN_SCRIPT_SRC_PATH, "/tmp/xbuild-script",
				),
			},
			// run the script
			dukkha.TaskExecSpec{
				IgnoreError: false,
				Command:     append(append(runCmd, "/tmp/xbuild-script"), s.ScriptArgs...),
				UseShell:    options.UseShell(),
				ShellName:   options.ShellName(),
			},
			// override that script
			dukkha.TaskExecSpec{
				IgnoreError: false,
				Command: sliceutils.NewStrings(
					options.ToolCmd(), "copy", "--chmod", "0644",
					replace_XBUILD_CURRENT_CONTAINER_ID,
					replace_XBUILD_RUN_SCRIPT_SRC_REDACTED_PATH, "/tmp/xbuild-script",
				),
			},
		)
	default:
		return nil, fmt.Errorf("invalid empty run statement")
	}

	return steps, nil
}