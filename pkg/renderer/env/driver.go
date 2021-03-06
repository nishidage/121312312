package env

import (
	"fmt"

	"arhat.dev/pkg/yamlhelper"
	"arhat.dev/rs"

	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/renderer"
	"arhat.dev/dukkha/pkg/templateutils"
)

const (
	DefaultName = "env"
)

func init() { dukkha.RegisterRenderer(DefaultName, NewDefault) }

func NewDefault(name string) dukkha.Renderer { return &Driver{name: name} }

type Driver struct {
	rs.BaseField `yaml:"-"`

	renderer.BaseRenderer `yaml:",inline"`

	name string

	// EnableExec controls arbitrary command execution support when expanding env.
	//
	// if set to false, expanding env with shell evaluation (e.g. `$(do something)`)
	// will be skipped and formatted
	//
	// NOTE: shell evaluation of backqouted string is always skipped and formatted
	//
	// Defaults to `false`
	EnableExec *bool `yaml:"enable_exec"`
}

func (d *Driver) RenderYaml(
	rc dukkha.RenderingContext, rawData interface{}, _ []dukkha.RendererAttribute,
) ([]byte, error) {
	rawData, err := rs.NormalizeRawData(rawData)
	if err != nil {
		return nil, err
	}

	bytesToExpand, err := yamlhelper.ToYamlBytes(rawData)
	if err != nil {
		return nil, fmt.Errorf(
			"renderer.%s: unsupported input type %T: %w",
			d.name, rawData, err,
		)
	}

	enableExec := d.EnableExec != nil && *d.EnableExec
	ret, err := templateutils.ExpandEnv(rc, string(bytesToExpand), enableExec)
	if err != nil {
		return nil, fmt.Errorf("renderer.%s: %w", d.name, err)
	}

	return []byte(ret), nil
}
