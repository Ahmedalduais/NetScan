<template>
  <div class="flex-none px-3 py-1.5 bg-net-surface border-t border-net-border flex items-center gap-2 text-xs text-net-muted">
    <!-- Status text -->
    <svg v-if="hasPermissionError" class="w-3.5 h-3.5 text-net-warning flex-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
    </svg>
    <span class="truncate flex-1">{{ status }}</span>
  </div>
</template>

<script setup>
defineProps({
  status: String,
  hasPermissionError: Boolean,
})
</script>
