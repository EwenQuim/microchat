# MicroChat

A real-time chat application built with Go (Fuego) backend and a modern frontend.

## Project Structure

```
microchat/
├── app/                    # Frontend application
├── cmd/                    # Application entry points
│   ├── server/            # API server
│   └── cli/               # CLI tool
├── internal/              # Private application code
│   ├── handlers/          # HTTP request handlers
│   ├── models/            # Data models
│   ├── services/          # Business logic
│   ├── repository/        # Data storage
│   ├── middleware/        # HTTP middleware
│   └── config/            # Configuration
├── pkg/                   # Public packages
│   └── client/            # API client library
├── static/                # Built frontend files (served by API)
├── scripts/               # Build and deployment scripts
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Docker Compose configuration
└── Makefile              # Build automation
```

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 20+
- Docker (optional)

### Installation

1. Install dependencies:
```bash
make install
```

2. Copy environment variables:
```bash
cp .env.example .env
```

### Development

Run in development mode (frontend dev server + Go server):
```bash
make dev
```

### Production Build

Build everything:
```bash
make build
```

Run the server:
```bash
make run
```

Or build and run specific components:
```bash
make build-frontend
make build-server
make build-cli
```

## Docker Deployment

### Using Pre-built Image (GitHub Container Registry)

Pull and run the latest image:
```bash
docker pull ghcr.io/ewenquim/microchat:latest
docker run -p 8080:8080 ghcr.io/ewenquim/microchat:latest
```

### Building Locally

Build Docker image:
```bash
make docker-build
```

Run with docker-compose:
```bash
make docker-up
```

Stop:
```bash
make docker-down
```

## CLI Usage

The CLI tool allows you to interact with the chat API from the command line.

Send a message:
```bash
./bin/cli -cmd send -room general -user john -message "Hello, world!"
```

List messages in a room:
```bash
./bin/cli -cmd list -room general
```

List all rooms:
```bash
./bin/cli -cmd rooms
```

## API Endpoints

- `GET /api/rooms` - List all chat rooms
- `GET /api/rooms/:room/messages` - Get messages from a room
- `POST /api/rooms/:room/messages` - Send a message to a room

## Environment Variables

- `PORT` - Server port (default: 8080)
- `ENV` - Environment (development/production)

## CI/CD

The project includes GitHub Actions workflows for:

- **CI** (`ci.yml`): Runs on PRs and pushes
  - Tests Go code (server & CLI)
  - Builds frontend
  - Runs with race detection and coverage

- **Lint** (`lint.yml`): Go code linting with golangci-lint
  - Strict on PRs (blocks merge on issues)
  - Warnings only on main/master

- **Docker Publish** (`docker-publish.yml`): Builds and pushes to GHCR
  - Triggers on push to main/master or version tags
  - Automatic tagging: `latest`, `v1.0.0`, `v1.0`, `v1`
  - Published to: `ghcr.io/ewenquim/microchat`

### Releasing

To create a new release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

This will automatically build and push the Docker image with version tags.

## Clean Up

Remove build artifacts:
```bash
make clean
```
