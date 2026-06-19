<script setup lang="ts">
const tabs = ['Timeline', 'Console', 'Files', 'Diff'] as const
defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<template>
  <div class="mb-3 flex gap-1 rounded-2xl bg-panel p-1">
    <button
      v-for="tab in tabs"
      :key="tab"
      type="button"
      class="clickable min-h-10 flex-1 rounded-xl text-sm font-medium capitalize transition-opacity duration-150 hover:opacity-95 active:opacity-80"
      :class="modelValue === tab.toLowerCase() ? 'bg-accent text-white shadow-lg shadow-accent/20' : 'text-muted'"
      @click="$emit('update:modelValue', tab.toLowerCase())"
    >
      {{ tab }}
    </button>
  </div>
</template>
