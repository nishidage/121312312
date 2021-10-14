package transform

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	dukkha_test "arhat.dev/dukkha/pkg/dukkha/test"
)

func TestDriver_RenderYaml(t *testing.T) {
	tests := []struct {
		name     string
		specStr  string
		expected interface{}
	}{
		{
			name: "Op Template YAML number",
			specStr: `
value: "10.10000"
ops:
- template: "{{- fromYaml .Value -}}"
`,
			expected: "10.1",
		},
		{
			name: "Op Template YAML str",
			specStr: `
value: "10.10000"
ops:
- template: '{{- fromYaml (printf "%q" .Value) -}}'
`,
			expected: "10.10000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := &driver{}
			ret, err := d.RenderYaml(dukkha_test.NewTestContext(context.TODO()), test.specStr)
			assert.NoError(t, err)
			assert.EqualValues(t, test.expected, ret)
		})
	}
}
