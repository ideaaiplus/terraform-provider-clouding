ARG GO_VERSION=1.25
ARG GOLANGCI_LINT_VERSION=v1.64.0

# ── base: Go toolchain + system deps ─────────────────────────────────────────
FROM golang:${GO_VERSION}-alpine AS base
WORKDIR /src
RUN apk add --no-cache git ca-certificates

# ── deps: download modules (cached as long as go.mod/go.sum are unchanged) ───
FROM base AS deps
COPY go.mod go.sum ./
RUN go mod download

# ── dev: all tools; used with -v $(PWD):/src for local iteration ──────────────
FROM deps AS dev
ARG GOLANGCI_LINT_VERSION
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

# ── tidy: verify go.mod/go.sum are tidy ──────────────────────────────────────
FROM deps AS tidy
COPY go.mod go.sum /before/
COPY . .
RUN go mod tidy && \
    diff /before/go.mod go.mod && \
    diff /before/go.sum go.sum

# ── build: compile check ──────────────────────────────────────────────────────
FROM deps AS build
COPY . .
RUN go build -o /dev/null ./...

# ── test: unit tests ──────────────────────────────────────────────────────────
FROM deps AS test
COPY . .
RUN go test -count=1 ./...

# ── lint: static analysis ────────────────────────────────────────────────────
FROM dev AS lint
COPY . .
RUN golangci-lint run ./...

# ── check: all of the above in one shot ──────────────────────────────────────
FROM dev AS check
COPY go.mod go.sum /before/
COPY . .
RUN go mod tidy && \
    diff /before/go.mod go.mod && \
    diff /before/go.sum go.sum && \
    go build -o /dev/null ./... && \
    go test -count=1 ./... && \
    golangci-lint run ./...
