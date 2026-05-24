<template>
  <div class="flex-none px-3 py-2 bg-net-surface border-b border-net-border flex flex-wrap items-center gap-2">
    <!-- Scan Button -->
    <button
      @click="$emit('scan')"
      :disabled="scanning"
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded transition-colors"
      :class="scanning
        ? 'bg-net-accent/20 text-net-accent cursor-not-allowed'
        : 'bg-net-accent/10 text-net-accent hover:bg-net-accent/20 active:bg-net-accent/30'"
    >
      <svg v-if="scanning" class="w-4 h-4 animate-spin-slow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
      <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      {{ scanning ? $t('toolbar.scanning') : $t('toolbar.scan') }}
    </button>

    <!-- Filter Input -->
    <div class="relative flex-1 min-w-[140px] max-w-xs">
      <svg class="absolute top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-net-muted pointer-events-none" :class="locale === 'ar' ? 'right-2' : 'left-2'" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        :placeholder="$t('toolbar.filter')"
        @input="$emit('update:filter', $event.target.value)"
        class="w-full bg-net-bg border border-net-border rounded text-sm px-7 py-1.5 text-net-text placeholder-net-muted outline-none focus:border-net-accent transition-colors"
      />
    </div>

    <!-- Language Toggle -->
    <button
      @click="$emit('toggle-lang')"
      class="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded border border-net-border text-net-muted hover:text-net-text hover:border-net-text/30 transition-colors"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      {{ $t('toolbar.language') }}
    </button>
  </div>
</template>

<script setup>
defineProps({
  scanning: Boolean,
  locale: String,
})
defineEmits(['scan', 'toggle-lang', 'update:filter'])
</script>
