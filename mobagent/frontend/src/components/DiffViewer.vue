<script setup lang="ts">
defineProps<{ diff: string }>()

function parseLines(diff: string) {
  return diff.split('\n').map((line, i) => ({ line, no: i + 1 }))
}
</script>

<template>
  <div class="glass-strong overflow-hidden rounded-2xl">
    <div class="max-h-72 overflow-auto p-2 font-mono text-[11px] leading-5">
      <div
        v-for="row in parseLines(diff)"
        :key="row.no"
        class="flex gap-2 px-1"
        :class="{
          'bg-success/12 text-success': row.line.startsWith('+') && !row.line.startsWith('+++'),
          'bg-danger/12 text-danger': row.line.startsWith('-') && !row.line.startsWith('---'),
          'text-accent-2': row.line.startsWith('@@'),
          'text-muted': !row.line.startsWith('+') && !row.line.startsWith('-') && !row.line.startsWith('@@'),
        }"
      >
        <span class="w-6 shrink-0 select-none text-right text-white/20">{{ row.no }}</span>
        <span class="min-w-0 flex-1 whitespace-pre-wrap break-all">{{ row.line || ' ' }}</span>
      </div>
    </div>
  </div>
</template>
