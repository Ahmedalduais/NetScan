<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-40" @click="$emit('close')"></div>
    <div v-if="visible"
      class="fixed z-50 bg-net-surface border border-net-border rounded-lg shadow-2xl py-1 min-w-[160px]"
      :style="{ left: x + 'px', top: y + 'px' }"
    >
      <button v-for="item in items" :key="item.label"
        @click.stop="onClick(item)"
        class="w-full text-left px-3 py-1.5 text-xs hover:bg-net-hover transition-colors flex items-center gap-2"
        :class="item.danger ? 'text-net-danger' : 'text-net-text'"
      >
        <span v-if="item.icon" class="text-[10px] w-4 text-center">{{ item.icon }}</span>
        {{ item.label }}
      </button>
    </div>
  </Teleport>
</template>

<script setup>
const props = defineProps({
  visible: { type: Boolean, default: false },
  x: { type: Number, default: 0 },
  y: { type: Number, default: 0 },
  items: { type: Array, default: () => [] },
})

const emit = defineEmits(['close', 'action'])

function onClick(item) {
  emit('action', item.action)
  emit('close')
}
</script>
