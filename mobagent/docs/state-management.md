# State Management

Pinia store: `stores/app.ts`

## State

- `agents`, `tasks` — REST loaded
- `events` — WebSocket + REST hydrate; capped at 100k IDs in Set dedupe
- `consoleLines`, `diffs` — per active task
- `activeTaskId`, `offline`, `loading`

## Actions

- `refresh()` — GET agents + tasks
- `selectTask(id)` — load events/console/diffs, connect WS
- `createTask()` — POST task, auto-select
- `taskAction(name)` — POST pause/resume/stop/…

## WebSocket

`WSClient` singleton per active task:
- reconnect 2s backoff
- `hello` + `lastEventId` resume
- client-side duplicate `id` filter

Frontend **never** parses terminal output into timeline events — only displays server `type`.
