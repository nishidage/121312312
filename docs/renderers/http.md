# HTTP Renderer

```yaml
foo@http: https://example.com/some-file
```

Render value using http

## Config Options

__NOTE:__ Configuration is required to activate this renderer.

```yaml
renderers:
- http:
    # cache config
    cache:
      # enable local cache, disable to always fetch from remote
      enabled: true
      timeout: 1h

    # http config
    method: GET # if not set, defaults to GET
    user: basic-auth-username
    password: basic-auth-password
    headers:
    - name: User-Agent
      value: dukkha
    # body: ""
    tls:
      enabled: false
      ca_cert: |-
        <pem-encoded-ca-cert>
      cert: |-
        <pem-encoded-cert>
      key: |-
        <pem-encoded-cert-key>
      server_name: server-name-override
      # key_log_file: for-tls-debugging
      # cipher_suites: []
      # insecure_skip_verify: true
    proxy:
      enabled: false
      http: http://proxy
      https: https://proxy
      # no_proxy:
      cgi: false
```

## Supported value types

- String: URL

  ```yaml
  foo@http: https://example.com/data
  ```

- Valid http fetch spec in yaml

  ```yaml
  foo@http:
    url: https://example.com/data
    config:
      # options are the same as Config Options .renderers.http
      # but without cache related options
      method: POST
  ```

## Supported Attributes

- `cached-file`: Return local file path to cached file instead of fetched content.

## Suggested Use Cases

- Organization to share dukkha config with certral http service.
- Download file via http
