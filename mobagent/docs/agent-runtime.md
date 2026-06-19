# Agent Runtime

## Start Task 流程

```
Vue NewTaskPage
  → POST /api/tasks  { prompt, agentId, workspace, branch, priority }
  → Runner.CreateTask
  → enqueue(agentId)     // 0=立即运行，>0=排队
  → adapter.Run()        // 模拟/真实 Agent 输出
  → parser → EventEngine → WebSocket → 前端 Timeline/Console
```

## Agent 复用与队列

- **每个 Agent 同一时刻只跑 1 个任务**（单 goroutine + 单 adapter 实例，资源极小）
- **同 Agent 多任务自动排队**（内存队列，无额外进程）
- 当前任务完成后自动拉起下一个
- 排队任务 `status=queued`，`queuePosition=N`
- `GET /api/agents/:id` 返回 `queueLength`

## 日志 (zap)

- 目录：`$LOG_DIR` 或 `data/logs/`
- 文件：`mobagent-api-YYYY-Www.log`（每周一个文件）
- 保留 30 天自动清理
- 全局 panic Recover + HTTP 请求日志

```bash
LOG_DIR=./data/logs
LOG_LEVEL=info
```
