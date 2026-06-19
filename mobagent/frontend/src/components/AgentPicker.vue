<script setup lang="ts">
import type { Agent } from '@/types'

defineProps<{
  modelValue: string
  agents: Agent[]
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const icon: Record<string, string> = {
  cursor: '⚡',
  'claude-code': '🧠',
  codex: '🤖',
  'gemini-cli': '✨',
  aider: '🔧',
  openhands: '🛠',
}

function select(id: string) {
  emit('update:modelValue', id)
  window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light')
}
</script>

<template>
  <div class="agent-picker" role="listbox" aria-label="Select agent">
    <button
      v-for="a in agents"
      :key="a.id"
      type="button"
      role="option"
      :aria-selected="modelValue === a.id"
      class="agent-picker__item clickable"
      :class="{ 'is-selected': modelValue === a.id }"
      @click="select(a.id)"
    >
      <span class="agent-picker__icon">{{ icon[a.type] ?? '⚡' }}</span>
      <span class="min-w-0 flex-1 text-left">
        <span class="block truncate text-sm font-semibold">{{ a.name }}</span>
        <span class="block truncate text-xs text-muted">{{ a.workspace }} · {{ a.status }}</span>
      </span>
      <span class="agent-picker__check" aria-hidden="true">{{ modelValue === a.id ? '✓' : '' }}</span>
    </button>
    <p v-if="!agents.length" class="rounded-xl bg-warning/10 px-3 py-2 text-center text-sm text-warning">
      No agents loaded — check backend connection
    </p>
  </div>
</template>

<style scoped>
.agent-picker {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.agent-picker__item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 0.75rem;
  min-height: 3.25rem;
  padding: 0.625rem 0.875rem;
  border-radius: 1rem;
  border: 1px solid color-mix(in srgb, white 8%, transparent);
  background: color-mix(in srgb, var(--color-panel) 90%, transparent);
  color: inherit;
  transition:
    border-color 0.15s ease,
    background-color 0.15s ease;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.agent-picker__item.is-selected {
  border-color: color-mix(in srgb, var(--color-accent) 55%, transparent);
  background: color-mix(in srgb, var(--color-accent) 12%, transparent);
}

.agent-picker__item:active {
  opacity: 0.85;
}

.agent-picker__icon {
  display: flex;
  height: 2.25rem;
  width: 2.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  border-radius: 0.75rem;
  background: color-mix(in srgb, white 6%, transparent);
  font-size: 1.125rem;
}

.agent-picker__check {
  flex-shrink: 0;
  width: 1.25rem;
  text-align: center;
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--color-accent);
}
</style>
