# File Renderer

```yaml
foo@file: path/to/local/file
```

Read local file content as the real value

## Config Options

```yaml
renderers:
- file:
    # enable in memory cache, or always read local file
    cache:
      enabled: true
      timeout: "0"
      size: 100M
    # set base path
    # if it's a relative path (without prefix `/`), will be joined with current working dir
    base_path: /somewhere
```

## Supported value types

- String: Local file path

  ```yaml
  foo@file: /tmp/data.json
  ```

- Any valid yaml value (only when `cache-data` attribute applied)

  ```yaml
  foo@file#cache-data: |-
    you can find me in DUKKHA_CACHE_DIR
  ```

## Supported Attributes

- `cache-data`: Save input data to cache, and return absolute local path to the cached file.
- `cached-file`: Return local path passed in directly (mainly intended to support offline mode of remote renderers)

## Suggested Use Cases

- Local config reuse
- Store content to file
