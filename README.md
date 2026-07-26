# App Repository — Proxyless gRPC Greeter Service

This repository contains the Go source code for a proxyless gRPC service mesh demonstration using Google Cloud Service Mesh (Traffic Director).

## Architecture

- **Server** (`cmd/server/`): A standard gRPC server implementing the `Greeter` service. Registered as a backend in Traffic Director via a Network Endpoint Group (NEG).
- **Client** (`cmd/client/`): A gRPC client that uses the `xds:///` scheme to connect to the server via Traffic Director for service discovery and load balancing — no sidecar proxy required.
- **Proto** (`helloworld/`): The protobuf definition and generated Go code for the `Greeter` service.

## How Proxyless gRPC Works

```
Client Pod                    Traffic Director (xDS)         Server Pod
─────────────────────────────────────────────────────────────────────
grpc.NewClient("xds:///greeter-service:50051")
    │
    ├─── xDS bootstrap config ──► Traffic Director
    │    (from init container)         │
    │                                  │ Returns: endpoint list,
    │◄─── Listener + RouteConfig ──────┘ load balancing policy
    │
    └─── gRPC call ──────────────────────────────────────────► :50051
         (load balanced across all server pods)
```

## Project Structure

```
app-repo/
├── Dockerfile                      # Multi-stage build (golang:1.23 → distroless)
├── go.mod                          # Go module definition
├── go.sum                          # Dependency checksums
├── cmd/
│   ├── server/
│   │   ├── server.go               # gRPC server implementation
│   │   └── server_test.go          # Server unit tests
│   └── client/
│       ├── client.go               # gRPC client implementation
│       └── client_test.go          # Client unit tests
└── helloworld/
    ├── helloworld.proto            # Protobuf service definition
    ├── helloworld.pb.go            # Generated message types
    ├── helloworld_grpc.pb.go       # Generated gRPC service code
    └── helloworld_test.go          # Protobuf descriptor validation tests
```

## Local Development

### Prerequisites

- Go 1.23+
- `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc` plugins

### Run Tests

```bash
go test -v ./...
```

### Build

```bash
go build -o server ./cmd/server
go build -o client ./cmd/client
```

### Regenerate Protobuf Code

If you modify `helloworld/helloworld.proto`, regenerate the Go code:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="$PATH:$(go env GOPATH)/bin"

protoc --proto_path=helloworld \
       --go_out=helloworld --go_opt=paths=source_relative \
       --go-grpc_out=helloworld --go-grpc_opt=paths=source_relative \
       helloworld/helloworld.proto
```

### Build Docker Image

```bash
docker build -t greeter:local .
```

## CI/CD Pipeline

The GitHub Actions workflow (`.github/workflows/ci.yaml`) runs on every push to `initial-setup`:

1. **Test** — runs `go mod tidy` and `go test -v ./...`
2. **Build & Push** — builds the Docker image and pushes to Artifact Registry:
   `us-central1-docker.pkg.dev/utila-eliran-home/greeter/greeter:<git-sha>`

### Required GitHub Secrets

| Secret | Description |
|--------|-------------|
| `GCP_PROJECT_NUMBER` | Numeric project number. Get with: `gcloud projects describe utila-eliran-home --format='value(projectNumber)'` |

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| `grpc.NewServer()` on server | Server-side xDS (`xds.NewGRPCServer()`) requires Traffic Director to push Listener resources, which creates a complex setup. The server is a standard gRPC backend; Traffic Director manages routing on the **client** side. |
| `xds:///` scheme on client | Enables proxyless service discovery and load balancing via Traffic Director without a sidecar proxy. |
| Separate health check registration | The gRPC health protocol is registered on the server so Traffic Director's health checks can verify pod readiness. |
| `cancel()` called explicitly in loop | Avoids context leak — `defer cancel()` inside a loop defers until function exit, not loop iteration. |
