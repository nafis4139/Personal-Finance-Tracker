# backend/Dockerfile

# ---------- Build Stage ----------
# Use an official Go toolchain image to compile the application.
# Multi-stage builds keep the final image small by discarding build-time dependencies.
FROM golang:1.25 AS builder
WORKDIR /app

# 1) Dependency metadata (cacheable layer).
# Copy go.mod and go.sum first so dependency downloads can be cached across builds.
COPY go.mod go.sum ./

# Optionally cache the module download directory for faster subsequent builds.
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# 2) Application source.
# Copy the remaining project files into the build context.
COPY . .

# 3) Reconcile dependencies.
# Ensures go.mod/go.sum are in sync with imports (useful when modules change).
RUN --mount=type=cache,target=/go/pkg/mod go mod tidy

# 4) Build the binary.
# - Disable CGO for a static binary suitable for distroless.
# - Target linux/amd64 explicitly.
# - -trimpath and stripped symbols (-s -w) reduce binary size.
# - Cache Go build artifacts to speed up iterative builds.
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /bin/api ./cmd/api

# ---------- Runtime Stage ----------
# Use a minimal distroless base for a smaller attack surface and reduced image size.
FROM gcr.io/distroless/base-debian12

# Default port exposed by the service (can be overridden at runtime).
ENV PORT=8080

# Work in the root directory to keep paths simple.
WORKDIR /

# Copy the compiled binary from the builder stage.
COPY --from=builder /bin/api /api

# Copy SQL migrations; expected to be discovered by the application at /migrations.
COPY migrations /migrations

# Document the listening port (optional metadata for tooling).
EXPOSE 8080

# Run as an unprivileged user for improved security.
USER nonroot:nonroot

# Entrypoint executes the compiled API server.
ENTRYPOINT ["/api"]
