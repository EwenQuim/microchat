# Claude Instructions

## Frontend Commands

Always use `npm run <script>` with the scripts defined in `package.json`. Never use `npx` or `pnpm`.

## API Client

Never write raw `fetch` calls for internal API requests. The API client is generated from the OpenAPI spec — use `npm run generate:api` (see `orval.config.ts`) to regenerate it after backend changes, then import from the generated files. Raw `fetch` is only acceptable for cross-origin requests to external/remote servers.

## TDD

Follow a test-driven development approach: write failing tests first, then implement the fix, then verify tests pass. For Go code, place tests in `_test.go` files in the same package.

## TUI Conventions

### Room Display Format
Rooms are displayed as `server~roomname`. The server prefix uses the server's quickname
if one is configured, otherwise it falls back to the hostname from the URL. The server
prefix is rendered dimmed to visually separate it from the room name.

### Navigation
Tab cycles through the four main sections: Rooms → Servers → Identities → Contacts → Rooms (wraps around).
Pressing Esc from any management screen (Servers, Identities, Contacts) always returns to Rooms.

### Multi-Server
The app connects to all configured servers simultaneously and shows rooms from all of them
in a single list. Rooms load in parallel; partial results appear immediately as each server
responds, with a loading indicator while others are still fetching.
