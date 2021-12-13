# filesystem of dukkha

We implemeneted in memory context aware local filesystem representation in `dukkha` to provide platform agnostic file access.

## Cache Files

Stored in `DUKKHA_CACHE_DIR`

- Intermediate files generated by tasks & renderers
- Files fetched from remote endpoints by renderers

## Special Files

__NOTE:__ Files mentioned below are only available in embedded bash environment, including renderer `shell`, `shell` action in hooks and `workflow:run` jobs, template func `eval.Shell`

- `/dev/null`: block all read operations, and bypass all write operations

## Local Files

We recommend using relative paths as much as possible, but when you have to work with absolute path on windows platform, there are some differences from regular windows paths:

- `C:\foo` is equivalent to `/c/foo`
  - in contrast to win32 behavior, `/c/foo` was interpreted as relative path of current disk drive, and it will be resolved to something like `D:\c\foo` depending on your work dir.