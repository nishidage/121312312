package tpl

import (
	"testing"

	"arhat.dev/rs"
	"github.com/stretchr/testify/assert"

	"arhat.dev/dukkha/pkg/dukkha"
	dt "arhat.dev/dukkha/pkg/dukkha/test"
)

var _ dukkha.Renderer = (*Driver)(nil)

func TestNewDriver(t *testing.T) {
	assert.NotNil(t, NewDefault(""))
}

func TestDriver_RenderYaml(t *testing.T) {
	dt.TestFixturesUsingRenderingSuffix(t, "./fixtures",
		map[string]dukkha.Renderer{
			"tpl": NewDefault("tpl"),
		},
		func() rs.Field { return &rs.AnyObjectMap{} },
		func() rs.Field { return &rs.AnyObjectMap{} },
		func(t *testing.T, ctx dukkha.Context, ts, cs rs.Field) {
			actual, expected := ts.(*rs.AnyObjectMap), cs.(*rs.AnyObjectMap)

			assert.EqualValues(t, expected.NormalizedValue(), actual.NormalizedValue())
		},
	)
}
