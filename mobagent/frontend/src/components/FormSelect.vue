<script setup lang="ts">
import { computed, ref } from 'vue'

export interface SelectOption {
  value: string
  label: string
  hint?: string
}

const props = defineProps<{
  modelValue: string
  options: SelectOption[]
  placeholder?: string
  title?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const open = ref(false)

const current = computed(() => props.options.find((o) => o.value === props.modelValue))

function openSheet() {
  open.value = true
}

function closeSheet() {
  open.value = false
}

function pick(value: string) {
  emit('update:modelValue', value)
  closeSheet()
  window.Telegram?.WebApp?.HapticFeedback?.impactOccurred('light')
}
</script>

<template>
  <div class="form-select">
    <button
      type="button"
      class="form-select__trigger glass clickable"
      :aria-expanded="open"
      @click.stop="openSheet"
      @touchend.prevent.stop="openSheet"
    >
      <span class="min-w-0 flex-1 text-left">
        <span v-if="current" class="block truncate text-sm text-white">{{ current.label }}</span>
        <span v-if="current?.hint" class="block truncate text-xs text-muted">{{ current.hint }}</span>
        <span v-else class="text-sm text-muted">{{ placeholder ?? 'Select…' }}</span>
      </span>
      <span class="form-select__chevron" aria-hidden="true">▾</span>
    </button>

    <Teleport to="body">
      <div
        v-if="open"
        class="form-select__overlay"
        @click.self="closeSheet"
        @touchend.self.prevent="closeSheet"
      >
        <div class="form-select__sheet sheet-safe-bottom" @click.stop @touchend.stop>
          <div class="mx-auto mb-3 h-1 w-10 rounded-full bg-white/20" />
          <h3 class="mb-3 text-center text-base font-semibold">{{ title ?? 'Choose' }}</h3>
          <div class="max-h-[50vh] space-y-1 overflow-y-auto">
            <button
              v-for="opt in options"
              :key="opt.value"
              type="button"
              class="form-select__option clickable"
              :class="{ 'is-active': opt.value === modelValue }"
              @click.stop="pick(opt.value)"
              @touchend.prevent.stop="pick(opt.value)"
            >
              <span class="block truncate font-medium">{{ opt.label }}</span>
              <span v-if="opt.hint" class="block truncate text-xs opacity-70">{{ opt.hint }}</span>
            </button>
            <p v-if="!options.length" class="py-6 text-center text-sm text-muted">No options</p>
          </div>
          <button type="button" class="form-select__cancel clickable" @click.stop="closeSheet">Cancel</button>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.form-select__trigger {
  display: flex;
  width: 100%;
  min-height: 3rem;
  align-items: center;
  gap: 0.75rem;
  border-radius: 1rem;
  padding: 0.625rem 0.875rem;
  text-align: left;
  touch-action: manipulation;
}

.form-select__chevron {
  flex-shrink: 0;
  color: color-mix(in srgb, white 45%, transparent);
  font-size: 1.125rem;
}

.form-select__overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  background: rgb(0 0 0 / 0.55);
  touch-action: manipulation;
}

.form-select__sheet {
  width: 100%;
  max-width: 32rem;
  border-radius: 1.25rem 1.25rem 0 0;
  padding: 1rem 1rem 1.25rem;
  background: color-mix(in srgb, var(--color-panel-2) 96%, black);
  border-top: 1px solid color-mix(in srgb, white 8%, transparent);
}

.form-select__option {
  width: 100%;
  border-radius: 0.875rem;
  padding: 0.875rem 1rem;
  text-align: left;
  background: transparent;
  color: inherit;
  touch-action: manipulation;
}

.form-select__option.is-active {
  background: color-mix(in srgb, var(--color-accent) 18%, transparent);
  color: var(--color-accent);
}

.form-select__cancel {
  margin-top: 0.75rem;
  width: 100%;
  min-height: 2.75rem;
  border-radius: 0.875rem;
  background: color-mix(in srgb, white 6%, transparent);
  color: color-mix(in srgb, white 55%, transparent);
  font-size: 0.875rem;
}
</style>
