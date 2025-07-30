# Custom Bridge for ConfigHub

This Go application implements a custom bridge worker for ConfigHub that wraps the standard Kubernetes bridge and adds file persistence functionality. It is based on the ConfigHub SDK examples and includes all standard functions.

## Features

- **Custom Kubernetes Bridge**: Wraps the standard Kubernetes bridge and adds file persistence
- **Standard Functions**: Includes all standard ConfigHub functions for Kubernetes YAML, OpenTofu HCL, and AppConfig Properties
- **File Persistence**: Saves config unit data and metadata to a configurable directory structure
- **Directory Structure**: Files are organized as `space-id/unit-slug/` with two files per unit:
  - `data.yaml`: Raw content of the "Data" field as readable YAML
  - `metadata.json`: JSON object with all config unit metadata (excluding Data field)

## Architecture

The application consists of:

1. **Main Application** (`main.go`): Sets up the worker with bridge dispatcher and function executor
2. **Custom Bridge** (`custom_bridge.go`): Wraps the standard Kubernetes bridge and adds file persistence
3. **Standard Functions**: All standard ConfigHub functions are automatically registered

## File Structure

```
custombridge/
├── main.go              # Main application entry point
├── custom_bridge.go     # Custom bridge implementation
├── go.mod              # Go module dependencies
├── spec.md             # Original specification
└── README.md           # This file
```

## Configuration

The application uses the following environment variables:

- `CONFIGHUB_WORKER_ID`: Your ConfigHub worker ID
- `CONFIGHUB_WORKER_SECRET`: Your ConfigHub worker secret
- `CONFIGHUB_URL`: ConfigHub server URL
- `CUSTOM_BRIDGE_DIR`: Base directory for file persistence (defaults to `/tmp/confighub-custom-bridge`)

## File Persistence Details

### On Apply Operation

When a config unit is applied:

1. The operation is delegated to the standard Kubernetes bridge
2. If successful, the config unit data and metadata are saved to files:
   - Directory: `{CUSTOM_BRIDGE_DIR}/{space-id}/{unit-slug}/`
   - `data.yaml`: Raw content of the "Data" field
   - `metadata.json`: JSON object with unit metadata (excluding Data field)

### On Destroy Operation

When a config unit is destroyed:

1. The operation is delegated to the standard Kubernetes bridge
2. If successful, the config unit files are deleted from the filesystem

### Directory Structure Example

```
/tmp/confighub-custom-bridge/
├── 550e8400-e29b-41d4-a716-446655440000/  # Space ID
│   ├── my-unit-1/
│   │   ├── data.yaml
│   │   └── metadata.json
│   └── my-unit-2/
│       ├── data.yaml
│       └── metadata.json
└── 6ba7b810-9dad-11d1-80b4-00c04fd430c8/  # Space ID
    └── another-unit/
        ├── data.yaml
        └── metadata.json
```

## Building and Running

### Prerequisites

- Go 1.24.3 or later
- Access to ConfigHub SDK

### Build

```bash
go mod tidy
go build -o custombridge
```

### Run

```bash
export CONFIGHUB_WORKER_ID="your-worker-id"
export CONFIGHUB_WORKER_SECRET="your-worker-secret"
export CONFIGHUB_URL="https://your-confighub-instance.com"
export CUSTOM_BRIDGE_DIR="/path/to/your/storage/directory"

./custombridge
```

## Development

This application is based on the ConfigHub SDK examples:

- **Bridge Example**: https://github.com/confighub/sdk/tree/main/examples/hello-world-bridge
- **Function Example**: https://github.com/confighub/sdk/tree/main/examples/hello-world-function

The custom bridge wraps the standard Kubernetes bridge implementation and adds file persistence functionality as specified in the requirements.

## License

MIT License - see the LICENSE file for details.
