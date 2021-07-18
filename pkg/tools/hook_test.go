package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	dukkha_test "arhat.dev/dukkha/pkg/dukkha/test"
	"arhat.dev/dukkha/pkg/field"
	"arhat.dev/dukkha/pkg/renderer/env"

	_ "embed"
)

var (
	//go:embed fixtures/001-hook-script-whitespace-trimed-after-rendering.yaml
	hookScriptWhitespaceTrimedAfterRendering []byte

	// nolint:revive
	//go:embed fixtures/001-expected.yaml
	_expected_001 []byte
)

func TestHookFixtures(t *testing.T) {
	testCases := []struct {
		name  string
		input []byte

		env      []string
		expected []byte
	}{
		{
			name:     "001-hook-script-whitespace-trimed-after-rendering",
			input:    hookScriptWhitespaceTrimedAfterRendering,
			env:      []string{"VERSION=1.26.1"},
			expected: _expected_001,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			ctx := dukkha_test.NewTestContext(context.TODO())
			ctx.AddRenderer("env", env.NewDefault(nil))
			ctx.AddEnv(test.env...)

			actual := field.Init(&Hook{}, nil).(*Hook)
			assert.NoError(t, yaml.Unmarshal(test.input, actual))
			assert.NoError(t, actual.ResolveFields(ctx, -1, ""))

			expected := field.Init(&Hook{}, nil).(*Hook)
			assert.NoError(t, yaml.Unmarshal(test.expected, expected))

			t.Log(actual)

			assert.EqualValues(t, expected.Cmd, actual.Cmd)
			assert.EqualValues(t, expected.ContinueOnError, actual.ContinueOnError)
			assert.EqualValues(t, expected.Other, actual.Other)
			assert.EqualValues(t, expected.Task, actual.Task)
		})
	}
}