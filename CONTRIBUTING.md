# Contributing to terraform-provider-clouding

Thank you for your interest in contributing! This guide covers everything you need to get started.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Releasing](#releasing)

---

## Quick start (Docker — no local toolchain needed)

If you have Docker installed you can run any check without installing Go or golangci-lint:

```bash
make docker-check    # tidy + build + test + lint in one shot
make docker-test     # unit tests only
make docker-lint     # static analysis only
make docker-tidy     # verify go.mod/go.sum are tidy
make docker-shell    # interactive shell with the repo mounted (for iteration)
```

The first run downloads modules and builds the toolchain image; subsequent runs reuse the Docker layer cache.

---

## Code of Conduct

Be respectful and constructive. We follow the
[Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

---

## Prerequisites

### Docker workflow (recommended)

| Tool | Version | Install |
|---|---|---|
| Docker | ≥ 24 | [docs.docker.com/get-docker](https://docs.docker.com/get-docker/) |
| Terraform | ≥ 1.0 | [developer.hashicorp.com/terraform/downloads](https://developer.hashicorp.com/terraform/downloads) (only for manual `terraform plan` runs) |

Go and golangci-lint run inside the container — no local installation required.

### Native workflow (optional — if you prefer to run tools directly)

| Tool | Version | Install |
|---|---|---|
| Go | ≥ 1.25 | [golang.org/doc/install](https://golang.org/doc/install) |
| Terraform | ≥ 1.0 | [developer.hashicorp.com/terraform/downloads](https://developer.hashicorp.com/terraform/downloads) |
| golangci-lint | v1.64+ | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| GoReleaser | latest | [goreleaser.com/install](https://goreleaser.com/install/) (only for release testing) |

---

## Development Setup

### 1. Fork and clone

```bash
git clone https://github.com/ideaaiplus/terraform-provider-clouding.git
cd terraform-provider-clouding
```

### 2. Verify the setup

**Docker (recommended):**
```bash
make docker-check
```

**Native:**
```bash
go mod tidy
```

### 3. Build the provider binary

```bash
go build -o terraform-provider-clouding .
```

### 4. Configure Terraform dev override

Create or edit `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "ideaaiplus/clouding" = "/absolute/path/to/terraform-provider-clouding"
  }
  direct {}
}
```

With this in place, Terraform will load your local binary instead of downloading from the registry.
**Note:** `terraform init` is not required for the provider when using dev overrides.

### 5. Test with the example config

```bash
cd examples
TF_VAR_clouding_api_key=YOUR_KEY terraform plan
```

---

## Project Structure

| Path | Purpose |
|---|---|
| `main.go` | Entry point — calls `providerserver.Serve` |
| `internal/client/client.go` | Pure `net/http` API client for Clouding.io |
| `internal/provider/provider.go` | Provider schema + `Configure` |
| `internal/provider/server_resource.go` | `clouding_server` CRUD resource |
| `examples/main.tf` | HCL example for local testing |
| `docs/` | Terraform Registry documentation (Markdown) |

See `CLAUDE.md` for architectural rules and coding conventions.

---

## Making Changes

### Branching strategy

Branch from `main` using one of these prefixes:

```
feat/<description>    # new resource, attribute, or behaviour
fix/<description>     # bug fix
docs/<description>    # documentation only
chore/<description>   # dependency bumps, CI changes
refactor/<description>
```

### Commit messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short imperative description>

Body (optional): explain the WHY, not the what.
```

**Types:** `feat` · `fix` · `docs` · `test` · `chore` · `refactor` · `ci`

**Scopes:** `client` · `provider` · `server` · `docs` · `release`

**Examples:**
```
feat(server): add firewall_id attribute
fix(client): handle 429 rate-limit responses gracefully
docs: add import instructions to server resource page
chore: bump terraform-plugin-framework to v1.20.0
```

### Adding a new resource

1. Create `internal/provider/<name>_resource.go` following the pattern in `server_resource.go`.
2. Implement `resource.Resource`, `resource.ResourceWithConfigure`, and `resource.ResourceWithImportState`.
3. Add compile-time interface assertions at the top of the file.
4. Register the factory in `internal/provider/provider.go → Resources()`.
5. Add the corresponding API methods to `internal/client/client.go`.
6. Create `docs/resources/<name>.md` following the template in `docs/resources/server.md`.
7. Add a usage example in `examples/resources/<name>/main.tf`.

### Dependency policy

- **No third-party HTTP libraries.** The `internal/client` package uses only `net/http`.
- New dependencies require a clear justification in the PR description.
- Run `go mod tidy` before every commit.

---

## Testing

### Docker (recommended — no local toolchain needed)

```bash
make docker-check   # all checks in one shot
make docker-test    # unit tests only
make docker-lint    # static analysis only
make docker-tidy    # go.mod/go.sum hygiene
make docker-shell   # interactive shell for iterative work
```

### Native (requires Go + golangci-lint installed)

```bash
make check          # all checks (tidy + build + test + lint)
make test           # unit tests only
make lint           # static analysis only
make tidy           # go.mod/go.sum hygiene
```

Or run individual commands directly:

```bash
go test ./...
go build -o /dev/null ./...
golangci-lint run ./...
go mod tidy && git diff --exit-code go.mod go.sum
```

### Acceptance tests (requires a real API key)

Acceptance tests hit the real Clouding.io API and create/destroy actual resources.
Set `TF_ACC=1` and `CLOUDING_API_KEY` before running:

```bash
export TF_ACC=1
export CLOUDING_API_KEY=your_key_here
go test ./internal/provider/... -v -run TestAcc
```

> **Warning:** acceptance tests create real infrastructure and may incur costs.
> Always verify resources are destroyed after the test run.

---

## Submitting a Pull Request

1. Ensure all checks pass locally (`go build`, `go test`, `golangci-lint`).
2. Push your branch and open a PR against `main`.
3. Fill in the PR template:
   - **What changed** — a concise description.
   - **Why** — the motivation or issue reference.
   - **How to test** — steps to verify the change.
4. A maintainer will review within a few business days.
5. Squash-merge is preferred to keep `main` history clean.

---

## Releasing

Releases are triggered by maintainers only, by pushing a semver tag to `main`:

```bash
git tag v0.2.0
git push origin v0.2.0
```

The `release.yml` GitHub Actions workflow then:
1. Builds binaries for Linux, macOS, Windows, FreeBSD (amd64 · arm64 · 386 · armv6).
2. Generates `SHA256SUMS` and signs it with the project GPG key.
3. Creates the GitHub release with all artifacts.
4. The Terraform Registry automatically indexes the new version.

**Do not create GitHub releases manually.**

---

## Questions?

Open a [GitHub Discussion](https://github.com/ideaaiplus/terraform-provider-clouding/discussions)
or file an [issue](https://github.com/ideaaiplus/terraform-provider-clouding/issues).
