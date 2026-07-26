# Stage 1: Build the Go binary
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the server and client
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /client ./cmd/client

# Stage 2: Create the final minimal image
FROM gcr.io/distroless/static-debian11

# Copy the binaries from the builder stage
COPY --from=builder /server /server
COPY --from=builder /client /client

# The server will be the default entrypoint
ENTRYPOINT ["/server"]
