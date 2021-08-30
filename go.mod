module arhat.dev/dukkha

go 1.16

require (
	arhat.dev/pkg v0.6.2
	arhat.dev/rs v0.1.3
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/aoldershaw/ansi v0.0.0-20210128170437-8c5426635e02
	github.com/die-net/lrucache v0.0.0-20210801000212-e34e67316dc5
	github.com/google/uuid v1.3.0
	github.com/gosimple/slug v1.10.0
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/huandu/xstrings v1.3.2
	github.com/itchyny/gojq v0.12.4
	github.com/minio/minio-go/v7 v7.0.12
	github.com/muesli/termenv v0.9.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	go.uber.org/multierr v1.7.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a
	golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	mvdan.cc/sh/v3 v3.3.1
)

replace github.com/creack/pty => github.com/jeffreystoke/pty v1.1.12-0.20210829095153-afe7d20a1998

replace arhat.dev/rs => arhat.dev/rs v0.1.3
