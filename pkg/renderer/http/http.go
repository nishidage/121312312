package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"arhat.dev/pkg/fshelper"
	"arhat.dev/pkg/rshelper"
	"arhat.dev/pkg/yamlhelper"
	"arhat.dev/rs"
	"gopkg.in/yaml.v3"

	"arhat.dev/dukkha/pkg/cache"
	"arhat.dev/dukkha/pkg/dukkha"
	"arhat.dev/dukkha/pkg/renderer"
)

const (
	DefaultName = "http"
)

func init() { dukkha.RegisterRenderer(DefaultName, NewDefault) }

func NewDefault(name string) dukkha.Renderer {
	return &Driver{name: name}
}

var _ dukkha.Renderer = (*Driver)(nil)

type Driver struct {
	rs.BaseField `yaml:"-"`

	renderer.BaseTwoTierCachedRenderer `yaml:",inline"`

	name string

	DefaultConfig rendererHTTPConfig `yaml:",inline"`

	defaultClient *http.Client
}

func (d *Driver) Init(cacheFS *fshelper.OSFS) error {
	err := d.BaseTwoTierCachedRenderer.Init(cacheFS)
	if err != nil {
		return err
	}

	d.defaultClient, err = d.DefaultConfig.createClient()
	return err
}

func (d *Driver) RenderYaml(
	rc dukkha.RenderingContext, rawData interface{}, attributes []dukkha.RendererAttribute,
) ([]byte, error) {
	var (
		reqURL string
		client *http.Client
		config *rendererHTTPConfig
	)

	rawData, err := rs.NormalizeRawData(rawData)
	if err != nil {
		return nil, err
	}

	switch t := rawData.(type) {
	case string:
		reqURL = t
		client = d.defaultClient
		config = &d.DefaultConfig
	case []byte:
		reqURL = string(t)
		client = d.defaultClient
		config = &d.DefaultConfig
	default:
		var rawBytes []byte
		rawBytes, err = yamlhelper.ToYamlBytes(rawData)
		if err != nil {
			return nil, fmt.Errorf(
				"renderer.%s: unexpected non yaml input: %w",
				d.name, err,
			)
		}

		spec := rshelper.InitAll(&inputHTTPSpec{}, &rs.Options{
			InterfaceTypeHandler: rc,
		}).(*inputHTTPSpec)
		err = yaml.Unmarshal(rawBytes, spec)
		if err != nil {
			return nil, fmt.Errorf(
				"renderer.%s: unmarshal input spec: %w",
				d.name, err,
			)
		}

		err = spec.ResolveFields(rc, -1)
		if err != nil {
			return nil, fmt.Errorf(
				"renderer.%s: resolving input spec: %w",
				d.name, err,
			)
		}

		// config resolved

		reqURL = spec.URL
		client, err = spec.Config.createClient()
		if err != nil {
			return nil, fmt.Errorf(
				"renderer.%s: creating http client for spec: %w",
				d.name, err,
			)
		}

		config = &spec.Config
	}

	data, err := renderer.HandleRenderingRequestWithRemoteFetch(
		d.Cache,
		cache.IdentifiableString(reqURL),
		func(_ cache.IdentifiableObject) (io.ReadCloser, error) {
			return d.fetchRemote(client, reqURL, config)
		},
		d.Attributes(attributes),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"renderer.%s: fetching http content: %w",
			d.name, err,
		)
	}

	return data, err
}

func (d *Driver) fetchRemote(
	client *http.Client,
	targetURL string,
	config *rendererHTTPConfig,
) (io.ReadCloser, error) {
	var (
		req *http.Request
		err error
	)

	var body io.Reader
	if config.Body != nil {
		body = strings.NewReader(*config.Body)
	}

	method := strings.ToUpper(config.Method)
	if len(method) == 0 {
		method = http.MethodGet
	}

	if len(config.BaseURL) != 0 {
		var baseURL *url.URL
		baseURL, err = url.Parse(config.BaseURL)
		if err != nil {
			return nil, err
		}

		baseURL.Path = path.Join(baseURL.Path, targetURL)
		targetURL = baseURL.String()
	}

	req, err = http.NewRequest(method, targetURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	if len(config.User) != 0 {
		req.SetBasicAuth(config.User, config.Password)
	}

	seen := make(map[string]struct{})
	for _, h := range config.Headers {
		_, ok := seen[h.Name]
		if !ok {
			seen[h.Name] = struct{}{}
			req.Header.Set(h.Name, h.Value)
		} else {
			req.Header.Add(h.Name, h.Value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
