export type EventType =
  | 'TaskStarted'
  | 'Planning'
  | 'ReadingProject'
  | 'ReadingFile'
  | 'Searching'
  | 'Editing'
  | 'CreatingFile'
  | 'UpdatingFile'
  | 'RemovingFile'
  | 'RunningCommand'
  | 'InstallingDependencies'
  | 'Building'
  | 'RunningTests'
  | 'Reviewing'
  | 'GitCommit'
  | 'GitPush'
  | 'NeedUserInput'
  | 'Warning'
  | 'Error'
  | 'Finished'
  | 'Failed'
  | 'DiffUpdated'
  | 'ConsoleOutput'
  | 'CostUpdated'
  | 'ResourceUpdated'
  | 'Heartbeat'

export interface AgentEvent {
  id: string
  taskId: string
  sessionId: string
  timestamp: string
  agent: string
  workspace: string
  type: EventType
  status: string
  payload?: Record<string, unknown>
}

export interface Agent {
  id: string
  name: string
  type: string
  status: string
  workspace: string
  currentStage?: string
  currentFile?: string
  progress: number
  elapsedSec: number
  tokens: number
  costUsd: number
  branch?: string
  cpu: number
  memoryMb: number
  currentTask?: string
  taskId?: string
}

export interface Task {
  id: string
  agentId: string
  agentType: string
  title: string
  prompt: string
  workspace: string
  branch?: string
  priority: string
  status: string
  progress: number
  stage?: string
  currentFile?: string
  tokens: number
  costUsd: number
  elapsedSec: number
  queuePosition?: number
  updatedAt?: string
  createdAt?: string
}

export interface FileDiff {
  path: string
  additions: number
  deletions: number
  status: string
  diff?: string
}

export interface TaskFile {
  path: string
  status: string
  additions?: number
  deletions?: number
}

export interface WSMessage<T = unknown> {
  type: string
  payload?: T
}
