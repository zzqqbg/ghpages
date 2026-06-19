<script setup lang="ts">
defineProps<{ placeholder?: string }>()
defineEmits<{ actions: [] }>()
</script>

<template>
  <div class="mt-4 flex gap-2">
    <input
      :placeholder="placeholder ?? 'Ask Agent...'"
      class="glass min-h-12 flex-1 rounded-2xl px-4 text-sm outline-none focus:ring-2 focus:ring-accent/40"
    />
    <button type="button" class="flex h-12 w-12 items-center justify-center rounded-2xl bg-accent text-lg">➤</button>
    <button type="button" class="glass flex h-12 w-12 items-center justify-center rounded-2xl text-lg" @click="$emit('actions')">⋯</button>
  </div>
</template>
