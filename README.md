# vand

Go module vanity import server. Responds to `?go-get=1` requests with the appropriate `go-import` meta tag so `go get` can resolve custom import paths to their VCS roots.

## Configuration

Place a `vand.yaml` (or `vand.yml`) in the working directory, or pass `--config <path>`.

```yaml
server:
  addr: ":8080"

packages:
  # Root package: `example.com` itself. The empty key matches `GET /?go-get=1`,
  # and the go.mod is at the repository root.
  "":
    target: https://github.com/example/example

  # Ordinary package: `example.com/x/net`, go.mod at the repository root.
  x/net:
    target: https://github.com/example/net
    branch: master

  # Subdirectory module: `example.com/pretty` whose go.mod lives in the `pretty/`
  # directory of the same repository. `root` advertises the go-import prefix
  # (here the domain root, `example.com`) so `go get` derives the subdirectory
  # from the remaining path.
  pretty:
    target: https://github.com/example/example
    root: ""

  x/sync:
    target: https://github.com/example/sync
    branch: master
    hidden: true   # suppresses pkg.go.dev indexing
```

### `packages` fields

The key is the import path under the host (e.g. `x/net` → `<host>/x/net`). The
empty key `""` is the host itself (the root package). Matching is exact — a
request for an unlisted path returns `404`, so the root package never acts as a
catch-all.

| Field    | Type   | Description                                                                                                              |
| -------- | ------ | ------------------------------------------------------------------------------------------------------------------------ |
| `target` | string | VCS repository root URL                                                                                                   |
| `branch` | string | Default branch (used for source links)                                                                                   |
| `hidden` | bool   | If true, adds `<meta name="robots" content="noindex"/>`                                                                   |
| `root`   | string | go-import prefix to advertise when this module's go.mod lives in a subdirectory of `target`. Must be a path prefix of the key; the subdirectory is the key with `root` removed. Omit for repo-root modules. |

#### Root and subdirectory modules

The `go-import` meta tag has no field for a subdirectory — `go` computes it as
`module path − advertised prefix`. So the prefix, not a separate field, controls
where `go` looks for the go.mod:

| Key          | `root` | go-import content                       | go.mod location in `target` |
| ------------ | ------ | --------------------------------------- | --------------------------- |
| `""`         | —      | `example.com git <target>`                 | repository root             |
| `x/net`      | —      | `example.com/x/net git <target>`           | repository root             |
| `pretty`     | `""`   | `example.com git <target>`                 | `pretty/`                   |
| `svc/kamino` | `svc`  | `example.com/svc git <target>`             | `kamino/`                   |

## Examples

### Minimal

A single package. With no `server` block the address defaults to `:8080`.

```yaml
packages:
  cli:
    target: https://github.com/example/cli
```

```console
$ go get example.com/cli
```

### A whole namespace on one domain

Map several import paths, each to its own repository, under one host. Each key
is matched exactly, so only these three paths resolve — anything else is `404`.

```yaml
server:
  addr: ":8080"

packages:
  cli:
    target: https://github.com/example/cli
  proto:
    target: https://github.com/example/proto
  sdk:
    target: https://github.com/example/sdk
```

| Import path      | Resolves to                              |
| ---------------- | ---------------------------------------- |
| `example.com/cli`   | `github.com/example/cli`           |
| `example.com/proto` | `github.com/example/proto`         |
| `example.com/sdk`   | `github.com/example/sdk`        |

### Root package

Serve the bare domain `example.com` as an import path. The empty key answers
`GET /?go-get=1`; the go.mod is at the repository root.

```yaml
packages:
  "":
    target: https://github.com/example/example
```

```go
import "example.com"
```

### Monorepo with submodules in subdirectories

When one repository holds several modules (a `go.mod` in more than one
directory), the `go-import` tag still names only the repository — it has no
field for a subdirectory. `go` finds the subdirectory by subtracting the
advertised prefix from the module path. `root` is that prefix.

#### Modules under the domain root

Repository `github.com/example/example` with a module at the root and one in
each subdirectory:

```
example/                       # github.com/example/example
├── go.mod                     # module example.com
├── pretty/
│   └── go.mod                 # module example.com/pretty
└── otelx/
    └── go.mod                 # module example.com/otelx
```

Set `root: ""` on the subdirectory modules so the advertised prefix is the
domain root, `example.com`. `go` then reads the subdirectory off the rest of
the import path (`example.com/pretty` − `example.com` → `pretty/`):

```yaml
packages:
  "":                 # example.com        -> go.mod at repo root
    target: https://github.com/example/example
  pretty:             # example.com/pretty -> pretty/go.mod
    target: https://github.com/example/example
    root: ""
  otelx:              # example.com/otelx  -> otelx/go.mod
    target: https://github.com/example/example
    root: ""
```

```go
import "example.com"        // module at the repo root
import "example.com/pretty" // module in pretty/
import "example.com/otelx"  // module in otelx/
```

#### Modules under a shared sub-prefix

When the modules live one level deep under a shared path segment, point `target`
at their repository and set `root` to that segment. `root` is a prefix of the
**key** (the import path); it is unrelated to the repository name in `target`.
`go` reads the subdirectory off what is left after removing it (key − root), so
`svc/kamino` with `root: svc` resolves to the `kamino/` directory. Here the
repository root maps to `example.com/svc`, so nothing is served at
`example.com/svc` itself — only the leaves are registered:

```
services/                      # github.com/example/services
├── kamino/
│   └── go.mod                 # module example.com/svc/kamino
└── tipoca/
    └── go.mod                 # module example.com/svc/tipoca
```

```yaml
packages:
  svc/kamino:         # example.com/svc/kamino -> kamino/go.mod
    target: https://github.com/example/services
    root: svc
  svc/tipoca:         # example.com/svc/tipoca -> tipoca/go.mod
    target: https://github.com/example/services
    root: svc
```

```go
import "example.com/svc/kamino" // module in kamino/
import "example.com/svc/tipoca" // module in tipoca/
```

### Mirroring / aliasing upstream packages

Point a short, stable import path at someone else's repository.

```yaml
packages:
  x/net:
    target: https://github.com/example/net
  x/sync:
    target: https://github.com/example/sync
```

### Hiding a package from indexing

`hidden: true` adds `<meta name="robots" content="noindex"/>` so the package is
not listed on pkg.go.dev. It still resolves for `go get`.

```yaml
packages:
  internal/secret:
    target: https://github.com/example/secret
    hidden: true
```

### Everything together

```yaml
server:
  addr: ":8080"

packages:
  "":
    target: https://github.com/example/example
  pretty:
    target: https://github.com/example/example
    root: ""
  cli:
    target: https://github.com/example/cli
  svc/kamino:
    target: https://github.com/example/services
    root: svc
  x/net:
    target: https://github.com/example/net
  internal/secret:
    target: https://github.com/example/secret
    hidden: true
```

## Endpoints

| Path                   | Description                                                    |
| ---------------------- | ------------------------------------------------------------- |
| `GET /`                | Health check (without `?go-get=1`) — returns `200 OK`, no logging |
| `GET /?go-get=1`       | Vanity import response for the root package (the `""` entry)   |
| `GET /<path>?go-get=1` | Vanity import response for the given package path             |

Any request without `?go-get=1` (other than the health check), or for an unknown package, returns `404`.

## Testing locally

Run the server against a config file (the `--config` flag is global, so it comes
before the subcommand):

```console
$ vand --config vand.yaml serve
```

Simulate a `go get` request with `curl`. Use `--header 'Host: ...'` to stand in
for the vanity domain and add `?go-get=1`:

```console
$ curl -s --header 'Host: example.com' 'http://localhost:8080/pretty?go-get=1'
<!DOCTYPE html><html><head>...<meta name="go-import" content="example.com git https://github.com/example/example"/></head><body></body></html>

$ curl -s --header 'Host: example.com' 'http://localhost:8080/svc/kamino?go-get=1'
<!DOCTYPE html><html><head>...<meta name="go-import" content="example.com/svc git https://github.com/example/services"/></head><body></body></html>

# Unlisted path -> 404
$ curl -s -o /dev/null -w '%{http_code}\n' --header 'Host: example.com' 'http://localhost:8080/nope?go-get=1'
404
```

To print the effective configuration (defaults applied, validated) without
starting the server:

```console
$ vand --config vand.yaml config
```
