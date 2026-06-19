# WebSocket Protocol

## Endpoint

`GET /ws` — WebSocket upgrade

## Client messages

### hello
```json
{ "type": "hello", "payload": { "sessionId": "uuid", "lastEventId": "abc", "taskId": "optional" } }
```

### subscribe
```json
{ "type": "subscribe", "payload": { "taskId": "task-id" } }
```

### ping
```json
{ "type": "ping" }
```

## Server messages

### hello_ack
```json
{ "type": "hello_ack", "payload": { "sessionId": "uuid", "serverTs": 1710000000 } }
```

### resume
Sent after reconnect when `lastEventId` is provided:
```json
{ "type": "resume", "payload": { "taskId": "…", "events": […], "lastEventId": "…" } }
```

### event
```json
{ "type": "event", "payload": { "id": "…", "taskId": "…", "type": "Planning", … } }
```

### heartbeat
Every 10s from server; client may also ping every 10s.

## Reconnect

1. Close → client waits 2s → reconnect
2. Send `hello` with same `sessionId` + `lastEventId`
3. Server sends `resume` batch for gap fill
4. Live `event` stream continues

Duplicate `id` values are ignored on both server (publish) and client (store).
