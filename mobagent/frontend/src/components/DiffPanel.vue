<script setup lang="ts">
import type { FileDiff } from '@/types'
import { computed } from 'vue'
import DiffViewer from '@/components/DiffViewer.vue'

const props = defineProps<{ diffs: FileDiff[]; selected?: string }>()
const emit = defineEmits<{ select: [path: string] }>()

const selectedPath = computed({
  get: () => props.selected || props.diffs[0]?.path || '',
  set: (v: string) => emit('select', v),
})

const current = computed(() => props.diffs.find((d) => d.path === selectedPath.value))
</script>

<template>
  <div class="flex h-full min-h-[20rem] flex-col gap-3">
    <label class="block">
      <select
        v-model="selectedPath"
        class="glass min-h-11 w-full rounded-xl px-3 font-mono text-sm outline-none"
      >
        <option v-for="d in diffs" :key="d.path" :value="d.path">{{ d.path }}</option>
      </select>
    </label>

    <DiffViewer v-if="current?.diff" :diff="current.diff" />

    <div class="glass rounded-2xl p-3">
      <p class="mb-2 text-xs font-medium text-muted">Changed Files</p>
      <div class="space-y-1.5">
        <button
          v-for="d in diffs"
          :key="d.path"
          type="button"
          class="flex w-full items-center justify-between rounded-lg px-2 py-2 text-left font-mono text-xs transition-colors"
          :class="selectedPath === d.path ? 'bg-accent/15 text-white' : 'hover:bg-white/5 text-muted'"
          @click="selectedPath = d.path"
        >
          <span class="truncate">{{ d.path }}</span>
          <span class="shrink-0 pl-2">
            <span class="text-success">+{{ d.additions }}</span>
            <span class="text-danger"> -{{ d.deletions }}</span>
          </span>
        </button>
      </div>
    </div>
  </div>
</template>
