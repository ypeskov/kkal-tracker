# syntax=docker/dockerfile:1

# Stage 1: Build the frontend (runs on the build host: the output is platform-independent)
FROM --platform=$BUILDPLATFORM node:24-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./ 
RUN npm run build

# Stage 2: Build the Go backend (runs on the build host, cross-compiles for the target platform)
FROM --platform=$BUILDPLATFORM golang:1.27 AS builder
WORKDIR /src

# Enable Go build and module caches for faster incremental builds
COPY go.mod go.sum ./ 
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Copy frontend build
COPY --from=frontend /app/web/dist ./web/dist

# Build a static binary for Linux
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/kkal-tracker ./cmd/web/main.go

# Stage 3: Create the final image
FROM gcr.io/distroless/static-debian13:nonroot AS prod
WORKDIR /app

# Copy the Go binary and migrations (configuration comes from the environment, no .env file in the image)
COPY --from=builder /out/kkal-tracker /usr/local/bin/kkal-tracker
COPY migrations ./migrations

# Default port; can be overridden by SERVER_PORT env var
EXPOSE 8080
ENV SERVER_PORT=8080
ENV ENVIRONMENT=production

# Run as non-root user provided by the base image
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/kkal-tracker"]