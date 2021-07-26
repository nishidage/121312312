package buildah

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/field"
	"arhat.dev/dukkha/pkg/sliceutils"
	"arhat.dev/dukkha/pkg/tools"
)

const TaskKindPush = "push"

func init() {
	dukkha.RegisterTask(
		ToolKind, TaskKindPush,
		func(toolName string) dukkha.Task {
			t := &TaskPush{
				manifestCache: make(map[manifestCacheKey]manifestCacheValue),
			}
			t.InitBaseTask(ToolKind, dukkha.ToolName(toolName), TaskKindPush, t)
			return t
		},
	)
}

type manifestCacheKey struct {
	execID int
	name   string
}

type manifestCacheValue struct {
	subIndex int
	name     string

	opts dukkha.TaskMatrixExecOptions
}

type TaskPush struct {
	field.BaseField

	tools.BaseTask `yaml:",inline"`

	ImageNames []ImageNameSpec `yaml:"image_names"`

	manifestCache map[manifestCacheKey]manifestCacheValue
}

func (c *TaskPush) cacheManifestPushSpec(
	opts dukkha.TaskMatrixExecOptions,
	manifestName string,
	index int,
) {
	key := manifestCacheKey{
		execID: opts.ID(),
		name:   manifestName,
	}

	c.manifestCache[key] = manifestCacheValue{
		subIndex: index,

		name: manifestName,
		opts: opts,
	}
}

func (c *TaskPush) createManifestPushSpecsFromCache(
	rc dukkha.TaskExecContext, execID int,
) []dukkha.TaskExecSpec {
	var (
		values []manifestCacheValue
	)

	// filter manifests belong to this exec
	for k, v := range c.manifestCache {
		if k.execID != execID {
			continue
		}

		values = append(values, v)
	}

	// restore original order
	sort.SliceStable(values, func(i, j int) bool {
		less := values[i].opts.Seq() < values[j].opts.Seq()
		if less {
			return true
		}

		return values[i].subIndex < values[j].subIndex
	})

	// generate specs using original options
	var ret []dukkha.TaskExecSpec
	for _, v := range values {
		delete(c.manifestCache, manifestCacheKey{
			execID: v.opts.ID(),
			name:   v.name,
		})

		// buildah manifest push --all \
		//   <manifest-list-name> <transport>:<transport-details>
		ret = append(ret, dukkha.TaskExecSpec{
			Env: sliceutils.NewStrings(c.Env),
			Command: sliceutils.NewStrings(
				v.opts.ToolCmd(), "manifest", "push", "--all",
				getLocalManifestName(v.name),
				// TODO: support other destination
				"docker://"+v.name,
			),
			IgnoreError: false,
			UseShell:    v.opts.UseShell(),
			ShellName:   v.opts.ShellName(),
		})
	}

	return ret
}

func (c *TaskPush) GetExecSpecs(
	rc dukkha.TaskExecContext,
	opts dukkha.TaskMatrixExecOptions,
) ([]dukkha.TaskExecSpec, error) {
	var result []dukkha.TaskExecSpec

	err := c.DoAfterFieldsResolved(rc, -1, func() error {
		targets := c.ImageNames
		if len(targets) == 0 {
			targets = []ImageNameSpec{
				{
					Image:    c.TaskName,
					Manifest: "",
				},
			}
		}

		dukkhaCacheDir := rc.CacheDir()

		for i, spec := range targets {
			if len(spec.Image) != 0 {
				imageName := SetDefaultImageTagIfNoTagSet(rc, spec.Image)
				imageIDFile := GetImageIDFileForImageName(
					dukkhaCacheDir, imageName,
				)
				imageIDBytes, err := os.ReadFile(imageIDFile)
				if err != nil {
					return fmt.Errorf("image id file not found: %w", err)
				}

				result = append(result, dukkha.TaskExecSpec{
					Env: sliceutils.NewStrings(c.Env),
					Command: sliceutils.NewStrings(
						opts.ToolCmd(), "push",
						string(bytes.TrimSpace(imageIDBytes)),
						// TODO: support other destination
						"docker://"+imageName,
					),
					IgnoreError: false,
					UseShell:    opts.UseShell(),
					ShellName:   opts.ShellName(),
				})
			}

			if len(spec.Manifest) == 0 {
				continue
			}

			manifestName := SetDefaultManifestTagIfNoTagSet(rc, spec.Manifest)
			c.cacheManifestPushSpec(opts, manifestName, i)
		}

		// push all manifests at last
		if opts.IsLast() {
			result = append(result,
				c.createManifestPushSpecsFromCache(rc, opts.ID())...,
			)
		}

		return nil
	})

	return result, err
}
