# Debian is used to avoid compatibility issues when using cgo.
FROM golang:1.22.4-bookworm AS builder

WORKDIR /user/src/

# Install external dependencies.

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential \
      libsqlite3-dev \
 && rm -rf /var/lib/apt/lists/*

# Install Go dependencies.

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./proto/ ./proto/

# Build the application.

# CGO_ENABLED=1 is required by go-sqlite3.
RUN --mount=type=cache,target=/user/gocache \
    GOCACHE=/user/gocache CGO_ENABLED=1 \
    go build -o /user/bin/ ./cmd/...


FROM debian:bookworm-slim AS runner

WORKDIR /user/

# Install external dependencies.

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      ca-certificates \
      libsqlite3-0 \
 && rm -rf /var/lib/apt/lists/*

# Prepare the user environment.

ENV UID=1000 GID=1000 HOME=/user
ENV PATH="/user/bin:$PATH"
RUN groupadd --gid "$GID" user \
 && useradd --uid "$UID" --gid "$GID" --home-dir "$HOME" --shell /bin/bash user \
 && mkdir -p /user/bin \
 && mkdir -p /user/data \
 && chown -R "$UID:$GID" /user

VOLUME /user/data

# Copy the application.

COPY --from=builder /user/bin/ /user/bin/

# Run the application.

USER user:user
