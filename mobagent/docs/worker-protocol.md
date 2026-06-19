# Worker WebSocket Protocol

内网 Worker 通过 **出站 WebSocket** 连接云端 Control Plane，支持多账户 token 隔离。

## Endpoint

```
GET /ws/worker?token=<account-token>
Authorization: Bearer <account-token>
X-MobAgent-Token: <account-token>
```

Token 在 `data/accounts.json` 或环境变量 `MOBAGENT_TOKENS=userId:token,...` 配置。

## 连接流程

```
Worker                          Cloud
  |---- WebSocket + token ------->|
  |<-------- auth_ok -------------|
  |-------- register ------------>|
  |<------- register_ok ----------|
  |<-------- run_task ------------|  (任务下发)
  |-------- event ---------------->|  (结构化事件)
  |-------- console -------------->|  (原始行)
  |-------- task_done ------------>|  (任务结束)
```

## Worker → Cloud

### register
```json
{
  "type": "register",
  "payload": {
    "workerId": "optional-stable-id",
    "hostname": "dev-box",
    "version": "1.0.0",
    "agents": [
      { "id": "cursor-1", "type": "cursor", "name": "Cursor Agent", "workspace": "mobagent", "branch": "main" }
    ],
    "workspaces": [
      { "name": "mobagent", "path": "/home/you/projs/mobagent" }
    ]
  }
}
```

### event
```json
{
  "type": "event",
  "payload": {
    "event": {
      "id": "...",
      "taskId": "...",
      "sessionId": "...",
      "type": "Editing",
      "status": "active",
      "agent": "cursor",
      "workspace": "mobagent",
      "payload": {}
    }
  }
}
```

### console
```json
{
  "type": "console",
  "payload": { "taskId": "...", "sessionId": "...", "line": "[12:01] Reading file" }
}
```

### task_done
```json
{
  "type": "task_done",
  "payload": { "taskId": "...", "sessionId": "...", "status": "completed", "error": "" }
}
```

## Cloud → Worker

### run_task
```json
{
  "type": "run_task",
  "payload": {
    "taskId": "...",
    "sessionId": "...",
    "agentId": "cursor-1",
    "agentType": "cursor",
    "prompt": "...",
    "workspace": "mobagent",
    "branch": "main",
    "priority": "medium"
  }
}
```

### cancel_task
```json
{ "type": "cancel_task", "payload": { "taskId": "..." } }
```

## 多账户

- 每个账户独立 token
- Worker 注册后，该账户的 `/api/agents` 和 `/api/workspaces` 仅返回该 Worker 上报的 catalog
- REST API 同样用 `Authorization: Bearer <token>` 隔离数据
- 未带 token 时默认 `demo` 账户（`MOBAGENT_REQUIRE_AUTH=1` 可强制认证）

## 本地 Worker 启动

```bash
export MOBAGENT_TOKEN=demo-token-change-me
export MOBAGENT_CLOUD_WS=ws://your-cloud:8790/ws/worker
export DATA_DIR=./data
export WORKSPACES_ROOT=./data/workspaces
go run ./cmd/worker/
# 或
./r worker
```

## 性能设计

- Worker send buffer 4096，UI buffer 256
- 独立 Worker Hub，不与 Telegram UI WebSocket 争锁
- Worker 侧 ping 30s，服务端 pong 90s 超时
- 断线指数退避重连（1s → 30s）
