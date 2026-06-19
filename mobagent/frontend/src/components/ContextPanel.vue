<script setup lang="ts">
import type { Agent } from '@/types'
import type { Task } from '@/types'
import { formatCost, formatDuration, formatTokens } from '@/lib/api'

defineProps<{
  agent?: Agent
  task?: Task
}>()
</script>

<template>
  <div class="glass mb-4 rounded-2xl p-3">
    <p class="mb-2 text-xs font-medium uppercase tracking-wide text-muted">Current Context</p>
    <div class="grid grid-cols-2 gap-3 text-xs">
      <div>
        <span class="text-muted">Workspace</span>
        <p class="mt-0.5 truncate font-medium">{{ agent?.workspace || task?.workspace || '-' }}</p>
      </div>
      <div>
        <span class="text-muted">Branch</span>
        <p class="mt-0.5 truncate font-medium">{{ agent?.branch || task?.branch || '-' }}</p>
      </div>
      <div>
        <span class="text-muted">Current File</span>
        <p class="mt-0.5 truncate font-mono text-accent-2">{{ task?.currentFile || agent?.currentFile || '-' }}</p>
      </div>
      <div>
        <span class="text-muted">Elapsed</span>
        <p class="mt-0.5 font-medium">{{ formatDuration(task?.elapsedSec ?? agent?.elapsedSec ?? 0) }}</p>
      </div>
      <div>
        <span class="text-muted">CPU</span>
        <p class="mt-0.5 font-medium">{{ agent?.cpu ?? 0 }}%</p>
      </div>
      <div>
        <span class="text-muted">Memory</span>
        <p class="mt-0.5 font-medium">{{ agent?.memoryMb ?? 0 }} MB</p>
      </div>
      <div>
        <span class="text-muted">Tokens</span>
        <p class="mt-0.5 font-medium">{{ formatTokens(task?.tokens ?? agent?.tokens ?? 0) }}</p>
      </div>
      <div>
        <span class="text-muted">Est. Cost</span>
        <p class="mt-0.5 font-medium">{{ formatCost(task?.costUsd ?? agent?.costUsd ?? 0) }}</p>
      </div>
    </div>
  </div>
</template>
