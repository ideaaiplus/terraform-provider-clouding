# CLAUDE.md — Project Intelligence for terraform-provider-clouding

This file is automatically loaded by Claude Code at the start of every session.
It defines the conventions, architecture, and rules every contributor (human or AI) must follow.

---

## Project Overview

A community Terraform provider for [Clouding.io](https://clouding.io), built with the
[HashiCorp Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework).

| Item | Value |
|---|---|
| Module | `github.com/ideaaiplus/terraform-provider-clouding` |
| Registry address | `registry.terraform.io/ideaaiplus/clouding` |
| API base URL | `https://api.clouding.io` |
| Auth header | `X-API-KEY` |
| Language | Go (see `go.mod` for exact version) |
| Framework | `github.com/hashicorp/terraform-plugin-framework` (latest) |
| License | MPL-2.0 |

---

## Repository Layout

```
.
├── main.go                              # Entry point — providerserver.Serve only
├── go.mod / go.sum
├── .goreleaser.yml                      # Multi-platform release config
├── .github/
│   └── workflows/
│       ├── test.yml                     # CI: build + unit tests on push/PR
│       └── release.yml                  # CD: GoReleaser on git tags
├── internal/
│   ├── client/
│   │   └── client.go                    # Pure net/http API client
│   └── provider/
│       ├── provider.go                  # Provider definition + Configure
│       └── server_resource.go           # clouding_server CRUD resource
├── examples/
│   └── main.tf                          # Local dev / acceptance test example
└── docs/
    ├── index.md                         # Provider docs (Terraform Registry)
    └── resources/
        └── server.md                    # clouding_server resource docs
```

---

## Architecture Rules

### 1. HTTP client — no third-party libraries
The `internal/client` package must use **only `net/http`** from the Go standard library.
Do not import `resty`, `retryablehttp`, `go-resty`, or any other HTTP wrapper.

### 2. Plugin Framework interfaces
Every resource must implement these interfaces (verified by compile-time assertions):
- `resource.Resource` — mandatory CRUD methods
- `resource.ResourceWithConfigure` — receives `*client.Client` from the provider
- `resource.ResourceWithImportState` — enables `terraform import`

Place compile-time assertions at the top of each resource file:
```go
var (
    _ resource.Resource                = &myResource{}
    _ resource.ResourceWithConfigure   = &myResource{}
    _ resource.ResourceWithImportState = &myResource{}
)
```

### 3. Provider data flow
`provider.Configure` builds a `*client.Client` and stores it in both
`resp.ResourceData` and `resp.DataSourceData`. Resources retrieve it in their
own `Configure` method via `req.ProviderData.(*client.Client)`.

### 4. State management
- `id` is always `Computed` + `UseStateForUnknown()` plan modifier.
- Fields that cannot be changed in-place (image, flavor, ssh key) use `RequiresReplace()`.
- Fields that can be updated in-place (name, hostname) have no plan modifier.
- In `Read`, a `*client.NotFoundError` always triggers `resp.State.RemoveResource(ctx)` — never an error.
- In `Delete`, a `*client.NotFoundError` is silently ignored (idempotent).

### 5. Error handling tiers
| Tier | Trigger | Action |
|---|---|---|
| Network / transport | `http.Client.Do` failure | Return raw `error` |
| 404 Not Found | API returns 404 | Return `*client.NotFoundError` |
| Other non-2xx | API returns 4xx/5xx | Return descriptive `fmt.Errorf` with status + body |

Resources call `resp.Diagnostics.AddError(summary, detail)` — never `panic`.

---

## Coding Conventions

### Go
- Follow standard Go formatting: run `gofmt` before committing (enforced by CI).
- Do not add comments explaining *what* the code does — names do that.
- Only add a comment when the **why** is non-obvious (hidden constraint, API quirk, workaround).
- All exported types and functions must have a GoDoc comment (one line max).
- No `init()` functions.
- Prefer `errors.As` over type assertions for error inspection.

### API field mapping
When the Clouding.io API uses camelCase JSON fields that differ from HCL snake_case names,
document the mapping in the struct tag and keep a note in `internal/client/client.go`:

| HCL attribute | API JSON field |
|---|---|
| `image_id` | `volume.id` (with `volume.source = "Image"`) |
| `flavor_id` | `flavorId` |
| `ssh_key_id` | `accessConfiguration.sshKeyId` |

### Commit messages — Conventional Commits
```
<type>(<scope>): <short description>

Types: feat | fix | docs | test | chore | refactor | ci
Scope: client | provider | server | docs | release (optional)

Examples:
  feat(server): add firewall_id attribute
  fix(client): handle 429 rate-limit with retry
  docs: update server resource import instructions
  chore: bump terraform-plugin-framework to v1.20.0
```

### Branch naming
```
feat/<short-description>
fix/<short-description>
docs/<short-description>
chore/<short-description>
```

---

## Adding a New Resource

1. Create `internal/provider/<name>_resource.go`.
2. Define the model struct with `tfsdk` tags matching the HCL schema.
3. Implement `resource.Resource`, `resource.ResourceWithConfigure`, `resource.ResourceWithImportState`.
4. Add compile-time assertions at the top.
5. Register the factory in `provider.go → Resources()`.
6. Add API methods to `internal/client/client.go`.
7. Create `docs/resources/<name>.md` following the template in `docs/resources/server.md`.
8. Add an example in `examples/resources/<name>/main.tf`.

---

## Build & Test Commands

### Docker (recommended — reproducible, no local toolchain required)

```bash
make docker-check   # tidy + build + test + lint in one shot
make docker-test    # unit tests only
make docker-lint    # static analysis only
make docker-tidy    # go.mod/go.sum hygiene
make docker-shell   # interactive shell with repo mounted
```

### Native (requires Go ≥ 1.25 + golangci-lint installed)

```bash
# Verify dependencies are tidy
go mod tidy && git diff --exit-code go.mod go.sum

# Compile check (no binary output)
go build -o /dev/null ./...

# Unit tests
go test ./...

# Build local binary for manual testing
go build -o terraform-provider-clouding .

# Lint (install golangci-lint first)
golangci-lint run ./...
```

### Local Terraform dev override

Add to `~/.terraformrc`:
```hcl
provider_installation {
  dev_overrides {
    "ideaaiplus/clouding" = "/home/labs/clouding-provider"
  }
  direct {}
}
```

Then test with `examples/main.tf`:
```bash
cd examples
TF_VAR_clouding_api_key=YOUR_KEY terraform plan
```

---

## Release Process

Releases are fully automated via GoReleaser + GitHub Actions.

1. Ensure `main` is clean and CI is green.
2. Create and push a semver tag:
   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```
3. The `release.yml` workflow fires: builds for 8 platforms, signs checksums with GPG,
   creates the GitHub release. The Terraform Registry picks it up automatically.

**Never manually edit GitHub releases** — GoReleaser owns them.

---

## What NOT to Do

- Do **not** use third-party HTTP libraries in `internal/client`.
- Do **not** hardcode API keys, tokens, or secrets anywhere in code or tests.
- Do **not** commit `*.tfvars`, `*.tfstate`, or the `dist/` folder.
- Do **not** skip `go mod tidy` before committing — CI will catch it and fail the PR.
- Do **not** write multi-line comment blocks. One-line GoDoc per exported symbol.
- Do **not** use `panic` anywhere in provider or client code.
- Do **not** create files in the repository root that are not listed in this document
  without updating this file and `CONTRIBUTING.md`.
