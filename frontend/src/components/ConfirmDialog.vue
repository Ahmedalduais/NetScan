<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="$emit('cancel')">
      <div class="bg-net-surface border border-net-border rounded-lg shadow-2xl max-w-sm w-full mx-4 p-5">
        <h3 class="text-sm font-semibold text-net-text mb-2">{{ title }}</h3>
        <p class="text-sm text-net-muted mb-5">{{ message }}</p>
        <div class="flex justify-end gap-2">
          <button @click="$emit('cancel')"
            class="px-3 py-1.5 text-xs rounded border border-net-border text-net-muted hover:text-net-text hover:border-net-text/30 transition-colors"
          >{{ cancelText }}</button>
          <button @click="$emit('confirm')"
            class="px-3 py-1.5 text-xs rounded transition-colors"
            :class="danger
              ? 'bg-net-danger text-white hover:bg-red-600'
              : 'bg-net-accent text-black hover:bg-cyan-400'"
          >{{ confirmText }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
defineProps({
  visible: { type: Boolean, default: false },
  title: { type: String, default: 'Confirm' },
  message: { type: String, default: 'Are you sure?' },
  confirmText: { type: String, default: 'Confirm' },
  cancelText: { type: String, default: 'Cancel' },
  danger: { type: Boolean, default: false },
})

defineEmits(['confirm', 'cancel'])
</script>
