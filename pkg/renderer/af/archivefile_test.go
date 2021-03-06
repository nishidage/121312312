package af

import (
	"context"
	"strings"
	"testing"

	"arhat.dev/pkg/testhelper"
	"arhat.dev/rs"
	"github.com/stretchr/testify/assert"

	di "arhat.dev/dukkha/internal"
	"arhat.dev/dukkha/pkg/dukkha"
	dukkha_test "arhat.dev/dukkha/pkg/dukkha/test"
	"arhat.dev/dukkha/pkg/renderer/file"
	"arhat.dev/dukkha/pkg/renderer/tpl"
)

func TestParseOneLineSpec(t *testing.T) {
	for _, test := range []struct {
		line     string
		expected *inputSpec
	}{
		{
			"foo:/bar", &inputSpec{
				Archive: "foo",
				Path:    "/bar",
			}},
		{
			"foo:bar", &inputSpec{
				Archive: "foo",
				Path:    "bar",
			}},
		{
			"foo:", &inputSpec{
				Archive: "foo",
				Path:    ".",
			},
		},
		{
			"foo", &inputSpec{
				Archive: "foo",
				Path:    "",
			},
		},
	} {
		t.Run(test.line, func(t *testing.T) {
			spec := parseOneLineSpec(test.line)
			assert.EqualValues(t, test.expected, spec)
		})
	}
}

func TestDriver(t *testing.T) {
	testhelper.TestFixtures(t, "./fixtures",
		func() interface{} {
			return rs.Init(&rs.AnyObject{}, nil)
		},
		func() interface{} {
			return rs.Init(&rs.AnyObject{}, nil)
		},
		func(t *testing.T, src, exp interface{}) {
			defer t.Cleanup(func() {})

			in := src.(*rs.AnyObject)
			check := exp.(*rs.AnyObject)

			ctx := dukkha_test.NewTestContext(context.TODO())
			ctx.(di.CacheDirSetter).SetCacheDir(t.TempDir())

			ctx.AddRenderer("file", file.NewDefault("file"))
			ctx.AddRenderer("tpl", tpl.NewDefault("tpl"))
			rdr := NewDefault("af")
			assert.NoError(t, rdr.Init(ctx.RendererCacheFS("af")))
			ctx.AddRenderer("af", rdr)
			ctx.AddEnv(true, &dukkha.EnvEntry{
				Name: "test_archives",
				Value: strings.Join([]string{
					// tar
					"001.tar",
					"002.tar.gz",
					"003.tar.bz2",
					"004.tar.lzma",
					"005.tar.xz",

					// zip
					"101.zip",
					"102.zip.gz",
					"103.zip.bz2",
					"104.zip.lzma",
					"105.zip.xz",
				}, " "),
			})

			assert.NoError(t, in.ResolveFields(ctx, -1))
			assert.NoError(t, check.ResolveFields(ctx, -1))

			actual := in.NormalizedValue()
			expected := check.NormalizedValue()

			assert.IsType(t, map[string]interface{}{}, expected)
			assert.IsType(t, expected, actual)

			assert.EqualValues(t, expected, actual)
		},
	)
}
