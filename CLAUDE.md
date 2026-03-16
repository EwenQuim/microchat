# Claude Instructions

## Frontend Commands

Always use `npm run <script>` with the scripts defined in `package.json`. Never use `npx` or `pnpm`.

## API Client

Never write raw `fetch` calls for internal API requests. The API client is generated from the OpenAPI spec — use `npm run generate:api` (see `orval.config.ts`) to regenerate it after backend changes, then import from the generated files. Raw `fetch` is only acceptable for cross-origin requests to external/remote servers.
