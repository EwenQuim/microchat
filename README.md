# MicroChat

A lightweight, self-hosted real-time chat app with Nostr-style cryptographic authentication. Built with Go and a modern web frontend.

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

### CLI

Send and read messages from the terminal:

```bash
# Send a message
./bin/cli -cmd send -room general -user john -message "Hello, world!"

# List messages in a room
./bin/cli -cmd list -room general

# List all rooms
./bin/cli -cmd rooms
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

**Project layout:** `app/` (frontend), `cmd/` (server + CLI entry points), `internal/` (handlers, services, models), `pkg/client/` (API client library).

**Releasing:** push a version tag — the Docker image is built and published to `ghcr.io/ewenquim/microchat` automatically.

```bash
git tag v1.0.0 && git push origin v1.0.0
```
