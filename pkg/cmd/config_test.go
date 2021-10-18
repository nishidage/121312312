package cmd

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"arhat.dev/dukkha/pkg/conf"
)

func TestReadConfigRecursively(t *testing.T) {
	var (
		testIncludeEmptyConfig = newConfig(func(c *conf.Config) {
			c.Include = []string{
				"empty.yaml",
				"empty-dir",
				// include self is ok, but will be ignored
				"include.yaml",
			}
		})
	)
	tests := []struct {
		name string

		configPaths        []string
		ignoreFileNotExist bool
		expected           *conf.Config
		expectErr          bool
	}{
		{
			name:     "None",
			expected: conf.NewConfig(),
		},

		// missing ok (ignoreFileNotExist=true)
		{
			name:               "Single File Missing OK",
			configPaths:        []string{"config-missing.yaml"},
			ignoreFileNotExist: true,
			expected:           conf.NewConfig(),
		},
		{
			name:               "Single Dir Missing OK",
			configPaths:        []string{"dir-missing"},
			ignoreFileNotExist: true,
			expected:           conf.NewConfig(),
		},
		{
			name:               "Multiple Missing OK",
			configPaths:        []string{"dir-missing", "config-missing.yaml"},
			ignoreFileNotExist: true,
			expected:           conf.NewConfig(),
		},

		// missing not ok (ignoreFileNotExist=false)
		{
			name:               "Single File Missing",
			configPaths:        []string{"config-missing.yaml"},
			ignoreFileNotExist: false,
			expectErr:          true,
		},
		{
			name:               "Single Dir Missing OK",
			configPaths:        []string{"dir-missing"},
			ignoreFileNotExist: false,
			expectErr:          true,
		},
		{
			name:               "Multiple Missing OK",
			configPaths:        []string{"dir-missing", "config-missing.yaml"},
			ignoreFileNotExist: false,
			expectErr:          true,
		},

		// empty
		{
			name:               "Empty Single File",
			configPaths:        []string{"empty.yaml"},
			ignoreFileNotExist: false,
			expected:           conf.NewConfig(),
		},
		{
			name:               "Empty Single Dir",
			configPaths:        []string{"empty-dir"},
			ignoreFileNotExist: false,
			expected:           conf.NewConfig(),
		},
		{
			name:               "Empty Multiple Source",
			configPaths:        []string{"empty-dir", "empty.yaml"},
			ignoreFileNotExist: false,
			expected:           testIncludeEmptyConfig,
		},
	}

	testFS := fstest.MapFS{
		"empty.yaml": &fstest.MapFile{
			Data: nil,
		},
		"empty-dir": &fstest.MapFile{
			Data: nil,
			Mode: fs.ModeDir,
		},
		"include-empty.yaml": &fstest.MapFile{
			Data: configBytes(t, testIncludeEmptyConfig),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			visitedPaths := make(map[string]struct{})
			mergedConfig := conf.NewConfig()

			err := readConfigRecursively(
				testFS,
				test.configPaths,
				test.ignoreFileNotExist,
				&visitedPaths,
				mergedConfig,
			)
			if test.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			expectedBytes, err := yaml.Marshal(test.expected)
			assert.NoError(t, err, "failed to marshal expected config")

			// field include is not merged
			mergedConfig.Include = test.expected.Include
			actualBytes, err := yaml.Marshal(mergedConfig)
			assert.NoError(t, err, "failed to marshal config loaded")

			assert.EqualValues(t, string(expectedBytes), string(actualBytes))
			t.Log(string(actualBytes))
		})
	}
}

func newConfig(update func(c *conf.Config)) *conf.Config {
	ret := conf.NewConfig()
	if update != nil {
		update(ret)
	}

	return ret
}

func configBytes(t *testing.T, c *conf.Config) []byte {
	data, err := yaml.Marshal(c)
	assert.NoError(t, err, "")
	return data
}