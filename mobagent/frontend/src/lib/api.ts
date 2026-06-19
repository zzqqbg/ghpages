const base = import.meta.env.VITE_API_BASE ?? ''

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function parseBody(res: Response): Promise<unknown> {
  const text = await res.text()
  if (!text) return null
  try {
    return JSON.parse(text) as unknown
  } catch {
    return text
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  const body = await parseBody(res)
  if (!res.ok) {
    const msg =
      typeof body === 'object' && body && 'error' in body
        ? String((body as { error: string }).error)
        : typeof body === 'string'
          ? body
          : res.statusText
    throw new ApiError(msg || 'request failed', res.status)
  }
  return body as T
}

export const api = {
  agents: () => request<import('@/types').Agent[]>('/api/agents'),
  agent: (id: string) => request<import('@/types').Agent>(`/api/agents/${id}`),
  tasks: () => request<import('@/types').Task[]>('/api/tasks'),
  createTask: (body: Record<string, string>) =>
    request<import('@/types').Task>('/api/tasks', { method: 'POST', body: JSON.stringify(body) }),
  taskAction: (id: string, action: string) =>
    request<import('@/types').Task>(`/api/tasks/${id}/${action}`, { method: 'POST' }),
  events: (id: string) => request<import('@/types').AgentEvent[]>(`/api/tasks/${id}/events`),
  console: (id: string) => request<{ lines: string[] }>(`/api/tasks/${id}/console`),
  diffs: (id: string) => request<import('@/types').FileDiff[]>(`/api/tasks/${id}/diffs`),
  files: (id: string) => request<import('@/types').TaskFile[]>(`/api/tasks/${id}/files`),
  workspaces: () => request<string[]>('/api/workspaces'),
}

export function formatDuration(sec: number) {
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}m ${s}s`
}

export function formatCost(usd: number) {
  return `$${usd.toFixed(3)}`
}

export function formatTokens(n: number) {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}
