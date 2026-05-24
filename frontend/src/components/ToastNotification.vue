<template>
  <Teleport to="body">
    <div class="fixed top-4 right-4 z-50 flex flex-col gap-2 max-w-sm" :class="locale === 'ar' ? 'left-4 right-auto' : 'right-4'">
      <div v-for="(toast, idx) in toasts" :key="idx"
        class="px-4 py-3 rounded-lg shadow-lg border text-sm flex items-start gap-3 animate-slide-in"
        :class="toastClass(toast.type)"
        style="animation: slideIn 0.25s ease-out"
      >
        <span class="mt-0.5">{{ toastIcon(toast.type) }}</span>
        <span class="flex-1">{{ toast.message }}</span>
        <button @click="$emit('dismiss', toast.id)" class="text-current opacity-50 hover:opacity-100 ml-2">&times;</button>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
const props = defineProps({
  toasts: { type: Array, default: () => [] },
  locale: { type: String, default: 'en' },
})

defineEmits(['dismiss'])

function toastClass(type) {
  switch (type) {
    case 'success': return 'bg-green-900/90 text-green-200 border-green-700'
    case 'error': return 'bg-red-900/90 text-red-200 border-red-700'
    case 'warning': return 'bg-yellow-900/90 text-yellow-200 border-yellow-700'
    case 'info': return 'bg-blue-900/90 text-blue-200 border-blue-700'
    default: return 'bg-net-surface text-net-text border-net-border'
  }
}

function toastIcon(type) {
  switch (type) {
    case 'success': return '✓'
    case 'error': return '✗'
    case 'warning': return '⚠'
    case 'info': return 'ℹ'
    default: return ''
  }
}
</script>

<style>
@keyframes slideIn {
  from { transform: translateX(100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
</style>
