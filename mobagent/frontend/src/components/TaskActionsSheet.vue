<script setup lang="ts">
defineProps<{ open: boolean }>()
defineEmits<{ close: []; action: [name: string] }>()

const items = [
  { key: 'pause', label: 'Pause Task', icon: '⏸' },
  { key: 'stop', label: 'Stop Task', icon: '■' },
  { key: 'restart', label: 'Restart Task', icon: '↺' },
  { key: 'logs', label: 'View Logs', icon: '📄' },
]
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-[100] flex items-end bg-black/65" @click.self="$emit('close')">
      <div class="glass-strong sheet-safe-bottom w-full max-w-lg rounded-t-3xl p-4">
        <div class="mx-auto mb-4 h-1 w-10 rounded-full bg-white/20" />
        <h3 class="mb-4 text-center text-lg font-semibold">Task Actions</h3>
        <div class="space-y-2">
          <button
            v-for="item in items"
            :key="item.key"
            type="button"
            class="flex min-h-14 w-full items-center gap-3 rounded-2xl bg-panel-2 px-4 text-left text-base"
            @click="$emit('action', item.key)"
          >
            <span class="text-xl">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </button>
        </div>
        <button type="button" class="mt-4 min-h-12 w-full rounded-2xl bg-white/5 text-muted" @click="$emit('close')">Cancel</button>
      </div>
    </div>
  </Teleport>
</template>
