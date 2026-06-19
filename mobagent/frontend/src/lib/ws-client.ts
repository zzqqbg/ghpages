import type { AgentEvent, WSMessage } from '@/types'

type Handler = (msg: WSMessage) => void

export class WSClient {
  private ws: WebSocket | null = null
  private url: string
  private sessionId: string
  private lastEventId = ''
  private taskId = ''
  private handlers = new Set<Handler>()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private pingTimer: ReturnType<typeof setInterval> | null = null
  private offline = false

  constructor(url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`) {
    this.url = url
    this.sessionId = crypto.randomUUID()
  }

  onMessage(fn: Handler) {
    this.handlers.add(fn)
    return () => this.handlers.delete(fn)
  }

  connect(taskId?: string) {
    if (taskId) this.taskId = taskId
    this.ws?.close()
    this.ws = new WebSocket(this.url)
    this.ws.onopen = () => {
      this.offline = false
      this.send({ type: 'hello', payload: { sessionId: this.sessionId, lastEventId: this.lastEventId, taskId: this.taskId } })
      if (this.taskId) this.send({ type: 'subscribe', payload: { taskId: this.taskId } })
      this.startPing()
    }
    this.ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data as string) as WSMessage
      if (msg.type === 'event' || msg.type === 'resume') {
        const events = msg.type === 'resume'
          ? (msg.payload as { events: AgentEvent[] }).events
          : [msg.payload as AgentEvent]
        for (const e of events) {
          if (e.id === this.lastEventId) continue
          this.lastEventId = e.id
          this.handlers.forEach((h) => h({ type: 'event', payload: e }))
        }
      } else {
        this.handlers.forEach((h) => h(msg))
      }
    }
    this.ws.onclose = () => {
      this.stopPing()
      this.offline = true
      this.handlers.forEach((h) => h({ type: 'offline' }))
      this.scheduleReconnect()
    }
    this.ws.onerror = () => this.ws?.close()
  }

  subscribe(taskId: string) {
    this.taskId = taskId
    this.send({ type: 'subscribe', payload: { taskId } })
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect(this.taskId)
    }, 2000)
  }

  private startPing() {
    this.stopPing()
    this.pingTimer = setInterval(() => this.send({ type: 'ping' }), 10000)
  }

  private stopPing() {
    if (this.pingTimer) clearInterval(this.pingTimer)
    this.pingTimer = null
  }

  private send(msg: WSMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  disconnect() {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.stopPing()
    this.ws?.close()
    this.ws = null
  }

  isOffline() {
    return this.offline
  }
}
