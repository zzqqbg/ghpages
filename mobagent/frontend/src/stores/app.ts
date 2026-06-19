import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Agent, AgentEvent, FileDiff, Task, TaskFile } from '@/types'
import { api } from '@/lib/api'
import { WSClient } from '@/lib/ws-client'

export const useAppStore = defineStore('app', () => {
  const agents = ref<Agent[]>([])
  const tasks = ref<Task[]>([])
  const events = ref<AgentEvent[]>([])
  const consoleLines = ref<string[]>([])
  const diffs = ref<FileDiff[]>([])
  const files = ref<TaskFile[]>([])
  const workspaces = ref<string[]>([])
  const activeTaskId = ref<string | null>(null)
  const offline = ref(false)
  const loading = ref(false)
  const lastError = ref<string | null>(null)
  const lastNotice = ref<string | null>(null)
  const loadError = ref<string | null>(null)
  const demoTaskId = ref('task-demo-1')

  let ws: WSClient | null = null
  const eventIds = new Set<string>()

  const runningAgents = computed(() => agents.value.filter((a) => a.status === 'running'))
  const activeTask = computed(() => tasks.value.find((t) => t.id === activeTaskId.value))

  async function loadAgents() {
    try {
      agents.value = await api.agents()
    } catch (e) {
      agents.value = []
      loadError.value = e instanceof Error ? e.message : 'Failed to load agents from server'
      throw e
    }
  }

  async function loadTasks() {
    try {
      tasks.value = await api.tasks()
    } catch {
      tasks.value = []
    }
  }

  async function loadWorkspaces() {
    try {
      workspaces.value = await api.workspaces()
    } catch (e) {
      workspaces.value = []
      loadError.value = e instanceof Error ? e.message : 'Failed to load workspaces from server'
      throw e
    }
  }

  async function loadBootstrap() {
    loadError.value = null
    try {
      await Promise.all([loadAgents(), loadWorkspaces(), loadTasks()])
    } catch {
      // loadError already set by failing loader
    }
  }

  function pushEvent(ev: AgentEvent) {
    if (eventIds.has(ev.id)) return
    eventIds.add(ev.id)
    events.value.push(ev)
    if (events.value.length > 100000) events.value.splice(0, events.value.length - 100000)
  }

  function connectWS(taskId?: string) {
    ws?.disconnect()
    ws = new WSClient()
    ws.onMessage((msg) => {
      if (msg.type === 'offline') offline.value = true
      if (msg.type === 'hello_ack' || msg.type === 'pong') offline.value = false
      if (msg.type === 'event') pushEvent(msg.payload as AgentEvent)
    })
    ws.connect(taskId)
  }

  async function selectTask(taskId: string) {
    activeTaskId.value = taskId
    try {
      const [evts, console, diffList, fileList] = await Promise.all([
        api.events(taskId),
        api.console(taskId),
        api.diffs(taskId),
        api.files(taskId),
      ])
      events.value = evts
      eventIds.clear()
      events.value.forEach((e) => eventIds.add(e.id))
      consoleLines.value = console.lines
      diffs.value = diffList
      files.value = fileList
    } catch {
      // keep prior state when API unavailable
    }
    connectWS(taskId)
  }

  async function createTask(payload: { prompt: string; agentId: string; workspace?: string; branch?: string; priority?: string }) {
    loading.value = true
    lastError.value = null
    lastNotice.value = null
    try {
      const task = await api.createTask(payload)
      tasks.value = [task, ...tasks.value.filter((t) => t.id !== task.id)]
      await selectTask(task.id)
      await loadAgents()
      if (task.status === 'queued') {
        lastNotice.value = `Task queued (#${task.queuePosition ?? '?'}) — agent is busy, will run next`
      } else {
        lastNotice.value = 'Task started — streaming events…'
      }
      return task
    } catch (e) {
      lastError.value = e instanceof Error ? e.message : 'Failed to start task'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function taskAction(action: string) {
    if (!activeTaskId.value) return
    const task = await api.taskAction(activeTaskId.value, action)
    const idx = tasks.value.findIndex((t) => t.id === task.id)
    if (idx >= 0) tasks.value[idx] = task
    else tasks.value.unshift(task)
    await loadAgents()
  }

  async function refresh() {
    loadError.value = null
    await Promise.all([loadAgents(), loadTasks()])
  }

  return {
    agents, tasks, events, consoleLines, diffs, files, workspaces,
    activeTaskId, offline, loading, demoTaskId, lastError, lastNotice, loadError,
    runningAgents, activeTask,
    loadAgents, loadTasks, loadWorkspaces, loadBootstrap, selectTask, createTask, taskAction, refresh, connectWS, pushEvent,
  }
})
