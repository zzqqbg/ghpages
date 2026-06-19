# Test Report

Generated: 2026-06-19

## Go (`backend/`)

| Suite | Tests | Status |
|-------|-------|--------|
| `internal/events` | dedupe, resume, ingest, 100k stress | PASS |
| `internal/parser` | dedupe reading, error detection | PASS |
| `internal/ws` | hello + event broadcast | PASS |

Command: `cd backend && go test ./...`

## Frontend Vitest

| File | Status |
|------|--------|
| `ws-client.test.ts` | PASS |
| `AgentCard.test.ts` | PASS |

Command: `cd frontend && pnpm test`

## E2E Playwright

| Spec | Status |
|------|--------|
| Home agents list | PASS (with API) |
| New task form | PASS |

Command: `cd frontend && pnpm exec playwright test`

Result: **2/2 PASS** (home agents, new task form)

## Performance

- Event engine stress: 100,000 events retained per task — PASS (`TestEngineStress100k`)
- Frontend event buffer capped at 100k with Set dedupe
- Console buffer capped at 10k lines server-side

## Accessibility

- Large touch targets (min-h-12 buttons)
- Dark theme default, semantic headings
- Focus rings on inputs (`focus:ring-accent/40`)

- TypeScript: `pnpm run build` PASS
- Go build: `go build ./cmd/server` PASS
- No placeholder UI routes — Home, Agent detail (Timeline/Console/Diff), New Task, Settings

## Remaining TODO

- Wire real Cursor/Claude Code CLI adapters (replace simulated runner)
- Telegram Bot auth (`initData` validation)
- Voice/image task composer attachments
