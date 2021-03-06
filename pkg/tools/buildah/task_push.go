package buildah

import (
	"bytes"
	"fmt"
	"sort"

	"arhat.dev/rs"

	"arhat.dev/dukkha/pkg/constant"
	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/templateutils"
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
			t.InitBaseTask(ToolKind, dukkha.ToolName(toolName), t)
			return t
		},
	)
}

type TaskPush struct {
	rs.BaseField `yaml:"-"`

	TaskName string `yaml:"name"`

	tools.BaseTask `yaml:",inline"`

	ImageNames []ImageNameSpec `yaml:"image_names"`

	manifestCache map[manifestCacheKey]manifestCacheValue
}

func (c *TaskPush) Kind() dukkha.TaskKind { return TaskKindPush }
func (c *TaskPush) Name() dukkha.TaskName { return dukkha.TaskName(c.TaskName) }
func (c *TaskPush) Key() dukkha.TaskKey {
	return dukkha.TaskKey{Kind: c.Kind(), Name: c.Name()}
}

func (c *TaskPush) GetExecSpecs(
	rc dukkha.TaskExecContext,
	opts dukkha.TaskMatrixExecOptions,
) ([]dukkha.TaskExecSpec, error) {
	var result []dukkha.TaskExecSpec

	err := c.DoAfterFieldsResolved(rc, -1, true, func() error {
		targets := c.ImageNames
		if len(targets) == 0 {
			targets = []ImageNameSpec{
				{
					Image:    c.TaskName,
					Manifest: "",
				},
			}
		}

		for i, spec := range targets {
			if len(spec.Image) != 0 {
				imageName := templateutils.SetDefaultImageTagIfNoTagSet(rc, spec.Image, true)
				imageIDFile, err := GetImageIDFileForImageName(rc, imageName, false)
				if err != nil {
					return err
				}

				imageIDBytes, err := rc.FS().ReadFile(imageIDFile)
				if err != nil {
					return fmt.Errorf("image id file not found: %w", err)
				}

				result = append(result, dukkha.TaskExecSpec{
					Command: []string{constant.DUKKHA_TOOL_CMD, "push",
						string(bytes.TrimSpace(imageIDBytes)),
						// TODO: support other destination
						"docker://" + imageName,
					},
					IgnoreError: false,
				})
			}

			if len(spec.Manifest) == 0 {
				continue
			}

			manifestName := templateutils.SetDefaultManifestTagIfNoTagSet(rc, spec.Manifest)
			c.cacheManifestPushSpec(i, opts, manifestName)
		}

		// push all manifests at last
		if opts.IsLast() {
			result = append(result,
				c.createManifestPushSpecsFromCache(opts.ID())...,
			)
		}

		return nil
	})

	return result, err
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

func (c *TaskPush) cacheManifestPushSpec(
	index int,
	opts dukkha.TaskMatrixExecOptions,
	manifestName string,
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

func (c *TaskPush) createManifestPushSpecsFromCache(execID int) []dukkha.TaskExecSpec {
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
	sort.Slice(values, func(i, j int) bool {
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
			Command: []string{constant.DUKKHA_TOOL_CMD, "manifest", "push", "--all",
				getLocalManifestName(v.name),
				// TODO: support other destination
				"docker://" + v.name,
			},
			IgnoreError: false,
		})
	}

	return ret
}
