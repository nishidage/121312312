# helm

## Supported Tasks

### Task `helm:package`

Package charts

```yaml
helm:package:
- name: example
  matrix:
    chart@shell: &charts |-
      for f in charts/* ; do
        echo "- charts/${f}"
      done
  # path to the chart
  chart@env: ${MATRIX_CHART}
  packages_dir: &pkg_dir .packages
  signing:
    enabled: false
    gpg_keyring: ${HOME}/.gnupg/secring.gpg
    gpg_key_name: <key-name>
    gpg_key_passphrase: <passphrase>
```

### Task `helm:index`

Update index file based on package files

```yaml
# this example is used in conjuction with the example for `helm:package`
helm:index:
- name: example
  repo_url: https://github.com/<org>/<repo>/releases/download/<github-release-name>/
  packages_dir: *pkg_dir
  # merge into the output
  merge: ./index.yaml
```
