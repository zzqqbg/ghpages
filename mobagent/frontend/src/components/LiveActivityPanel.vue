<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { formatCost, formatDuration, formatTokens } from '@/lib/api'

const store = useAppStore()
const router = useRouter()

const running = computed(() => store.agents.find((a) => a.status === 'running'))
const recentEvents = computed(() =>
  store.events
    .filter((e) => e.type !== 'ConsoleOutput' && e.type !== 'Heartbeat')
    .slice(-4)
    .reverse(),
)

const eventLabel: Record<string, string> = {
  TaskStarted: 'Task started',
  ReadingProject: 'Reading project',
  Searching: 'Searching codebase',
  Planning: 'Planning',
  Editing: 'Editing files',
  RunningTests: 'Running tests',
  Reviewing: 'Reviewing',
  Finished: 'Completed',
}

function openLive() {
  if (!running.value) return
  router.push({
    name: 'agent',
    params: { id: running.value.id },
    query: running.value.taskId ? { task: running.value.taskId } : {},
  })
}
</script>

<template>
  <section class="live-panel clickable" @click="openLive">
    <div class="live-panel__head">
      <div>
        <p class="live-panel__eyebrow">Live Activity</p>
        <h3 class="live-panel__title">{{ running?.currentStage || 'Waiting for agent…' }}</h3>
      </div>
      <span v-if="running" class="live-panel__pulse">Live</span>
    </div>

    <div v-if="running" class="live-panel__stats">
      <div class="live-panel__stat">
        <span class="label">Progress</span>
        <span class="value">{{ running.progress }}%</span>
      </div>
      <div class="live-panel__stat">
        <span class="label">Elapsed</span>
        <span class="value">{{ formatDuration(running.elapsedSec) }}</span>
      </div>
      <div class="live-panel__stat">
        <span class="label">Tokens</span>
        <span class="value">{{ formatTokens(running.tokens) }}</span>
      </div>
      <div class="live-panel__stat">
        <span class="label">Cost</span>
        <span class="value">{{ formatCost(running.costUsd) }}</span>
      </div>
    </div>

    <div v-if="running" class="live-panel__bar">
      <div class="live-panel__bar-fill" :style="{ width: `${running.progress}%` }" />
    </div>

    <ul v-if="recentEvents.length" class="live-panel__feed">
      <li v-for="ev in recentEvents" :key="ev.id" class="live-panel__feed-item">
        <span class="dot" />
        <span class="text">{{ eventLabel[ev.type] || ev.type }}</span>
      </li>
    </ul>

    <p v-else class="live-panel__empty">Start a task to see realtime events here</p>
  </section>
</template>

<style scoped>
.live-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 11rem;
  margin-bottom: 1rem;
  padding: 1rem;
  border-radius: 1rem;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-panel-2) 88%, transparent), color-mix(in srgb, var(--color-panel) 92%, transparent));
  border: 1px solid color-mix(in srgb, white 8%, transparent);
}

.live-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.live-panel__eyebrow {
  font-size: 0.6875rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: color-mix(in srgb, var(--color-accent) 80%, white);
}

.live-panel__title {
  margin-top: 0.25rem;
  font-size: 0.9375rem;
  font-weight: 600;
  line-height: 1.35;
}

.live-panel__pulse {
  flex-shrink: 0;
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  background: color-mix(in srgb, var(--color-success) 16%, transparent);
  color: var(--color-success);
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.live-panel__stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.5rem;
  margin-top: 0.875rem;
}

.live-panel__stat .label {
  display: block;
  font-size: 0.625rem;
  color: color-mix(in srgb, white 38%, transparent);
}

.live-panel__stat .value {
  display: block;
  margin-top: 0.125rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.live-panel__bar {
  height: 0.375rem;
  margin-top: 0.75rem;
  overflow: hidden;
  border-radius: 999px;
  background: color-mix(in srgb, white 8%, transparent);
}

.live-panel__bar-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--color-accent), var(--color-accent-2));
  transition: width 0.3s ease;
}

.live-panel__feed {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  gap: 0.5rem;
  margin-top: 0.875rem;
  padding-top: 0.875rem;
  border-top: 1px solid color-mix(in srgb, white 6%, transparent);
}

.live-panel__feed-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: color-mix(in srgb, white 72%, transparent);
}

.live-panel__feed-item .dot {
  width: 0.375rem;
  height: 0.375rem;
  border-radius: 999px;
  background: var(--color-accent);
  box-shadow: 0 0 8px color-mix(in srgb, var(--color-accent) 60%, transparent);
}

.live-panel__empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 1rem;
  font-size: 0.8125rem;
  color: color-mix(in srgb, white 34%, transparent);
  text-align: center;
}
</style>
