# Stage 1: Build the Go binaries
FROM golang:1.23-alpine AS builder

# Allow Go to automatically download a newer toolchain if required by go.mod
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Copy the Go module files and download dependencies first (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Tidy modules to ensure go.mod and go.sum are consistent with the toolchain
RUN go mod tidy

# Build the server and client as static binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /client ./cmd/client

# Stage 2: Create the final minimal image using distroless
FROM gcr.io/distroless/static-debian12

# Copy the binaries from the builder stage
COPY --from=builder /server /server
COPY --from=builder /client /client

# The server will be the default entrypoint
ENTRYPOINT ["/server"]
