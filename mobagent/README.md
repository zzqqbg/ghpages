# MobAgent

Telegram Mini App — 远程控制 Cursor / Claude Code / Codex 等 AI 编程代理。

## 快速开始

```bash
./r init
./r dev      # 后端 air 热重载 :8790 + 前端 Vite :5175
./r test     # Go + Vitest
./r b        # fzf 选择构建目标
./r p user@host   # 发布到 ~/projs/mobagent（OpenResty + systemd）
```

## 架构

```
Agent Adapter → Raw Output → Parser → Event Engine → WebSocket → Vue Mini App
```

前端只消费结构化 Event，原始终端输出仅在 Console 页展示。

详见 [docs/architecture.md](docs/architecture.md)
