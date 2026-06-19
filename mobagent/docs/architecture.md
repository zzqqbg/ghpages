# Architecture

## Overview

MobAgent is a Telegram Mini App for remotely controlling local AI coding agents with **event streaming** (not raw terminal UI on the home screen).

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│ Agent       │───▶│ Adapter      │───▶│ Parser      │───▶│ Event Engine │
│ (Cursor…)   │    │ (simulated)  │    │ (normalize) │    │ dedupe/merge │
└─────────────┘    └──────────────┘    └─────────────┘    └──────┬───────┘
                                                                  │
                                                                  ▼
┌─────────────┐    ┌──────────────┐                      ┌──────────────┐
│ Vue 3 UI    │◀───│ WebSocket    │◀─────────────────────│ Gin HTTP API │
│ Pinia store │    │ reconnect    │                      │ REST + static│
└─────────────┘    └──────────────┘                      └──────────────┘
```

## Backend packages

| Package | Role |
|---------|------|
| `internal/adapter` | Pluggable agent runners (Cursor, Claude Code, …) |
| `internal/parser` | Raw line → typed `AgentEvent` |
| `internal/events` | Dedupe, buffer (100k), broadcast hook |
| `internal/ws` | WebSocket hub, hello/resume/subscribe, heartbeat |
| `internal/agent` | Task lifecycle + actions |
| `internal/store` | In-memory agents/tasks/diffs/console |

## Event model

Every event: `id`, `taskId`, `sessionId`, `timestamp`, `agent`, `workspace`, `type`, `status`, `payload`.

Types include `TaskStarted`, `Planning`, `ReadingProject`, `Editing`, `RunningTests`, `DiffUpdated`, `ConsoleOutput`, `Finished`, `Heartbeat`, …

## WebSocket protocol

Client → Server:
- `hello` `{ sessionId, lastEventId?, taskId? }`
- `subscribe` `{ taskId }`
- `ping`

Server → Client:
- `hello_ack`, `resume` (missed events), `event`, `heartbeat`, `pong`

## Deployment

`./r p` rsyncs to `~/projs/mobagent`, installs user systemd unit, writes OpenResty `conf.d/mobagent.conf` (SPA + `/api` + `/ws` proxy). Idempotent re-runs.
