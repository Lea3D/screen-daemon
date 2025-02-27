# Define build arguments to allow easy version updates
ARG GOLANG_VERSION=${GOLANG_VERSION:-1.24}  # Allows runtime override for easier upgrades

# Use the specified Go version on a Debian Bookworm base image as the build dependencies stage
FROM golang:${GOLANG_VERSION}-bookworm AS build_deps

# Install necessary build dependencies
# Git is required for fetching dependencies, but it is only needed in the build stage
RUN apt-get update && apt-get install -y --no-install-recommends git && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /workspace

# Copy Go module dependency files first to leverage Docker's build cache and avoid unnecessary dependency downloads
COPY go.mod .
COPY go.sum .

# Download dependencies before copying the application source
RUN go mod download

# Build stage: Compile the application
FROM build_deps AS build

# Copy the entire source code
COPY . .

# Build the application as a statically linked binary
# Disabling CGO ensures portability across different systems
# The '-w' flag removes debugging information, reducing binary size
# The '-extldflags "-static"' ensures full static linking for portability
RUN CGO_ENABLED=0 go build -o mqtt2cmd -ldflags '-w -extldflags "-static"' .

# Final runtime image using Debian Bookworm Slim
FROM debian:bookworm-slim

# Install necessary runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Ensure the config directory exists and copy a default example config if no config exists
RUN mkdir -p /workspace/internal/config \
    && [ ! -f /workspace/internal/config/config.yaml ] && cp /workspace/internal/config/example_config.yaml /workspace/internal/config/config.yaml || true

# Set up a persistent volume for configuration files
VOLUME ["/workspace/internal/config"]

# Bind host directory for persistent configuration storage
RUN mkdir -p /workspace/internal/config \
    && ln -s /srv/docker/mqtt2cmd/config /workspace/internal/config

# Copy only the built binary from the build stage to keep the final image minimal
COPY --from=build /workspace/mqtt2cmd /usr/local/bin/mqtt2cmd

# Set the entry point to the application binary
ENTRYPOINT ["/usr/local/bin/mqtt2cmd"]
