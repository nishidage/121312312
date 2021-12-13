# Transform Renderer `T`

```yaml
foo@T:
  value: foo
  ops:
  - template: |-
      {{ VALUE }}-do-something
  - shell: |-
      echo ${VALUE}
```

Use operations to transform string value.

## Config Options

```yaml
renderers:
  # no options
- T: {}
```

## Supported value types

Valid transform spec in yaml

```yaml
foo@T:
  # value is a string value
  value: String Only, seriously
  # operations you want to take on the value
  ops:
  # Execute golang template over VALUE
  - template: |-
      add some {{- /* go */ -}} template
      your value above is available as {{ VALUE }}
  # Execute shell script with env ${VALUE}
  - shell: |-
      echo "${VALUE}"

  # Checksum verify data integrity, input value is returned as result
  - checksum:
      path@tpl: "{{ VALUE }}"
      # kind of the checksum, with
      kind: sha256
      # hex encoded sum value
      sum: # ...
      # key once set, verify hmac
      key: # ...
```

## Suggested Use Cases

- Convert your non-yaml data to yaml right in yaml.
- Composite different renderers to achieve significantly more flexibility.

  ```yaml
  # step (0): entrypoint is the `T` renderer
  # step (4): `af` renders value generated by renderer `T`
  foo@T|af:
    # (1) first step happens here: fetch data.tar.gz from remote http endpoint
    # notice the `#cached-file`, attribute `cached-file` will make renderer
    # `http` return local file path to the cached content.
    value@http#cached-file: https://example.com/data.tar.gz
    ops:
    # step (2): verify checksum of the downloaded archive
    - checksum:
        file@env: ${VALUE}
        kind: sha256
        sum: # sha256sum of the expected file
    # step (3): format the resolved `value` for render `af`
    # we are using type hint `str` to convert map as string since
    # template operation only accepts string value
    - template@?str:
        archive: {{ VALUE }}
        path: in-archive-target-file
  ```