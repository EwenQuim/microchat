# MicroChat

A lightweight, self-hosted real-time chat app with Nostr-style cryptographic authentication. Built with Go and a modern web frontend.

## Install the CLI

```bash
# Homebrew (macOS / Linux)
brew install ewenquim/repo/microchat

# Go
go install github.com/EwenQuim/microchat/cmd/microchat@latest
```

## Quick Start (Docker)

```bash
docker run -p 8080:8080 ghcr.io/ewenquim/microchat:latest
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Running from Source

**Prerequisites:** Go 1.23+, Node.js 20+

```bash
cp .env.example .env
make install
make dev
```

This starts the Go server and frontend dev server concurrently.

## Usage

### Web UI

Open the app in your browser, create or join a room, and start chatting. Authentication uses Nostr-style keypairs — no account registration required.

### Client (TUI + CLI)

Build from source:

```bash
make build-microchat   # outputs bin/microchat
```

Run without arguments to launch the interactive TUI:

```bash
microchat
```

Or use subcommands for scripting:

```bash
# Send a message
microchat send --room general --user john --message "Hello, world!"

# List messages in a room
microchat list --room general

# List all rooms
microchat rooms

# Connect to a different server
microchat --url http://chat.example.com rooms
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `ENV` | `development` | Environment (`development` / `production`) |

## Self-Hosting with Docker Compose

```bash
make docker-up
```

To stop:

```bash
make docker-down
```

## API

- `GET /api/rooms` — List all chat rooms
- `GET /api/rooms/:room/messages` — Get messages from a room
- `POST /api/rooms/:room/messages` — Send a message to a room

## Contributing

**Build & run:**

```bash
make build
make run
```

**Project layout:** `app/` (frontend), `cmd/` (server + microchat entry points), `internal/` (handlers, services, models, tui), `pkg/client/` (API client library).

**Releasing:** push a version tag — the Docker image is built and published to `ghcr.io/ewenquim/microchat` automatically.

```bash
git tag v1.0.0 && git push origin v1.0.0
```
