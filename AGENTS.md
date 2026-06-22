# Claude Instructions

## Frontend Commands

Always use `npm run <script>` with the scripts defined in `package.json`. Never use `npx` or `pnpm`.

## API Client

Never write raw `fetch` calls for internal API requests. The API client is generated from the OpenAPI spec — use `npm run generate:api` (see `orval.config.ts`) to regenerate it after backend changes, then import from the generated files. Raw `fetch` is only acceptable for cross-origin requests to external/remote servers.

## TDD

Follow a test-driven development approach: write failing tests first, then implement the fix, then verify tests pass. For Go code, place tests in `_test.go` files in the same package.

## Conventions

### Room Display Format

Rooms are displayed as `server~roomname`. The server prefix uses the server's quickname
if one is configured, otherwise it falls back to the hostname from the URL. The server
prefix is rendered dimmed to visually separate it from the room name.

### Navigation

The app is a single two-pane view: a left sidebar (rooms list, with `Servers`/`Identities`/`Contacts`
pinned at the bottom) and a right pane that shows either the selected room's chat or the selected
management section.

- Moving the cursor `↑↓` through the sidebar auto-previews the highlighted item in the right pane
  (a room shows its chat; a nav item shows that config). When a config section is shown, the chat
  title and message input are hidden.
- `S` / `I` / `C` jump the cursor to Servers / Identities / Contacts and preview them.
- `←`/`→` (or `Tab`) move focus between the sidebar and the right pane, exactly like the rooms↔chat
  switch. `Enter` on a nav item previews and focuses it in one step.
- From a focused config pane, `←`/`Esc`/`Tab` return focus to the sidebar. Config edits made in the
  right pane are saved immediately.

The full-screen Servers/Identities/Contacts screens remain only for first-run onboarding.

### Multi-Server

The app connects to all configured servers simultaneously and shows rooms from all of them
in a single list. Rooms load in parallel; partial results appear immediately as each server
responds, with a loading indicator while others are still fetching.
