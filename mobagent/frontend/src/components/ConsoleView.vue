<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'

const props = withDefaults(defineProps<{ lines: string[]; placeholder?: string }>(), {
  placeholder: 'Ask Agent or type command...',
})

const container = ref<HTMLElement | null>(null)
const paused = ref(false)
const query = ref('')

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.lines
  return props.lines.filter((l) => l.toLowerCase().includes(q))
})

function lineClass(line: string) {
  if (line.includes('PASS')) return 'rounded-lg bg-success/10 px-2 py-1 text-success'
  if (/error|fail|fatal/i.test(line)) return 'text-danger'
  if (/Test Suites:|Tests:/.test(line)) return 'text-accent-2 font-medium'
  if (/successfully|🎉/.test(line)) return 'text-success font-medium'
  return 'text-muted'
}

async function copyAll() {
  await navigator.clipboard.writeText(props.lines.join('\n'))
}

watch(() => props.lines.length, async () => {
  if (paused.value) return
  await nextTick()
  container.value?.scrollTo({ top: container.value.scrollHeight, behavior: 'smooth' })
})
</script>

<template>
  <div class="flex h-full min-h-[20rem] flex-col">
    <div class="mb-2 flex gap-2">
      <input
        v-model="query"
        placeholder="Search logs..."
        class="glass min-h-10 flex-1 rounded-xl px-3 text-sm outline-none focus:ring-2 focus:ring-accent/40"
      />
      <button type="button" class="glass min-h-10 rounded-xl px-3 text-sm" @click="paused = !paused">{{ paused ? 'Auto' : 'Pause' }}</button>
      <button type="button" class="glass min-h-10 rounded-xl px-3 text-sm" @click="copyAll">Copy</button>
    </div>
    <div ref="container" class="glass-strong flex-1 overflow-y-auto rounded-2xl p-3 font-mono text-[11px] leading-relaxed">
      <div
        v-for="(line, i) in filtered"
        :key="i"
        class="mb-1 whitespace-pre-wrap break-all"
        :class="lineClass(line)"
      >
        {{ line }}
      </div>
      <p v-if="!filtered.length" class="py-8 text-center text-sm text-muted">Waiting for output...</p>
    </div>
  </div>
</template>
