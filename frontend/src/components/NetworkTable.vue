<template>
  <div class="flex gap-3 h-full">
    <!-- Left Panel: Interface List + Blocked -->
    <div class="w-[300px] flex-none flex flex-col gap-2">
      <!-- Interfaces list -->
      <div class="flex-1 space-y-1 overflow-y-auto min-h-0">
        <div v-if="loading" class="flex items-center justify-center py-16">
          <svg class="w-6 h-6 text-net-accent animate-spin-slow" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </div>

        <div v-else-if="!interfaces || interfaces.length === 0" class="flex flex-col items-center justify-center py-16 text-net-muted">
          <svg class="w-10 h-10 mb-2 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.858 15.355-5.858 21.213 0" /></svg>
          <p class="text-xs">{{ $t('table.no_data') }}</p>
        </div>

        <template v-for="iface in interfaces" :key="iface.name">
          <button @click="selectView('iface', iface.name)"
            class="w-full text-left px-2.5 py-2 rounded border transition-all"
            :class="viewType === 'iface' && viewId === iface.name
              ? 'border-net-accent bg-net-accent/5'
              : 'border-transparent bg-net-surface hover:bg-net-hover hover:border-net-border'"
          >
            <div class="flex items-center gap-2">
              <span class="w-2 h-2 rounded-full flex-none" :class="iface.is_up ? 'bg-net-success' : 'bg-net-danger'"></span>
              <span class="font-mono text-xs font-medium flex-1 truncate">{{ iface.name }}</span>
              <span class="text-[9px] px-1 py-0.5 rounded font-medium"
                :class="iface.is_up ? 'bg-net-success/10 text-net-success' : 'bg-net-danger/10 text-net-danger'"
              >{{ iface.is_up ? $t('table.up') : $t('table.down') }}</span>
            </div>
            <div class="flex items-center gap-3 mt-1 text-[9px] text-net-muted">
              <span>{{ iface.ip_addresses.length }} IP</span>
              <span>{{ iface.connections.length }} conn</span>
            </div>
            <div v-if="tp(iface.name)" class="flex items-center gap-2 mt-0.5 text-[8px]">
              <span class="text-green-400/80">▼{{ tp(iface.name).rx_bps_str }}</span>
              <span class="text-blue-400/80">▲{{ tp(iface.name).tx_bps_str }}</span>
            </div>
          </button>
        </template>
      </div>

      <!-- Bottom Section: Blocked + Processes -->
      <div class="flex-none border-t border-net-border pt-2 space-y-1">
        <button @click="selectView('blocked', '')"
          class="w-full text-left px-2.5 py-2 rounded border transition-all"
          :class="viewType === 'blocked'
            ? 'border-net-danger bg-net-danger/5'
            : 'border-transparent bg-net-surface hover:bg-net-hover hover:border-net-border'"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm">🚫</span>
            <span class="text-xs font-medium flex-1">{{ $t('table.blocked') }}</span>
            <span class="text-[9px] px-1.5 py-0.5 rounded bg-net-danger/10 text-net-danger font-medium">{{ blockedCount }}</span>
          </div>
        </button>
        <button @click="selectView('processes', '')"
          class="w-full text-left px-2.5 py-2 rounded border transition-all"
          :class="viewType === 'processes'
            ? 'border-net-accent bg-net-accent/5'
            : 'border-transparent bg-net-surface hover:bg-net-hover hover:border-net-border'"
        >
          <div class="flex items-center gap-2">
            <span class="text-sm">⚙️</span>
            <span class="text-xs font-medium flex-1">{{ $t('table.processes') }}</span>
            <span class="text-[9px] px-1.5 py-0.5 rounded bg-net-accent/10 text-net-accent font-medium">{{ processes.length }}</span>
          </div>
        </button>
      </div>
    </div>

    <!-- Right Panel: Detail View -->
    <div class="flex-1 overflow-y-auto min-w-0">
      <!-- Blocked Items View -->
      <div v-if="viewType === 'blocked'" class="space-y-2 pr-1">
        <div class="flex items-center gap-2 text-sm font-semibold">
          <span>🚫</span>
          <span>{{ $t('table.blocked') }}</span>
          <span class="text-xs text-net-muted font-normal">({{ blockedCount }})</span>
        </div>

        <div v-if="blockedCount === 0" class="text-net-muted text-xs py-8 text-center">{{ $t('table.no_blocked') }}</div>

        <div v-else class="space-y-1">
          <!-- Blocked IPs -->
          <div v-if="blockedIPs.length > 0">
            <div class="text-[10px] text-net-muted mb-1 font-medium">IP Addresses</div>
            <div v-for="b in blockedIPs" :key="b.id"
              class="flex items-center gap-2 px-2.5 py-2 bg-net-surface rounded border border-red-900/30"
            >
              <span class="text-[9px] px-1 py-0.5 rounded bg-blue-500/10 text-blue-400 font-mono">IP</span>
              <span class="font-mono text-xs flex-1">{{ b.target }}</span>
              <span v-if="b.interface" class="text-[9px] text-net-muted truncate max-w-[80px]">{{ b.interface }}</span>
              <span class="text-[9px] px-1 py-0.5 rounded bg-net-danger/10 text-net-danger">BLOCKED</span>
              <button @click="$emit('unblock', b.id)"
                class="text-[9px] px-2 py-0.5 rounded bg-net-bg text-net-muted hover:text-net-text border border-net-border hover:border-net-text/30 transition-colors"
              >{{ $t('table.unblock') }}</button>
            </div>
          </div>

          <!-- Blocked PIDs -->
          <div v-if="blockedPIDs.length > 0">
            <div class="text-[10px] text-net-muted mb-1 mt-3 font-medium">{{ $t('table.processes') }} (PID)</div>
            <div v-for="b in blockedPIDs" :key="b.id"
              class="flex items-center gap-2 px-2.5 py-2 bg-net-surface rounded border border-red-900/30"
            >
              <span class="w-5 h-5 rounded flex-none flex items-center justify-center text-[8px] font-bold text-white shadow-sm"
                :style="{ backgroundColor: processColor(b.target.replace('PID ', '')) }"
              >{{ processInitial(b.target.replace('PID ', '')) }}</span>
              <span class="font-mono text-xs flex-1">{{ b.target }}</span>
              <span class="text-[9px] px-1 py-0.5 rounded bg-net-danger/10 text-net-danger">BLOCKED</span>
              <button @click="$emit('unblock', b.id)"
                class="text-[9px] px-2 py-0.5 rounded bg-net-bg text-net-muted hover:text-net-text border border-net-border hover:border-net-text/30 transition-colors"
              >{{ $t('table.unblock') }}</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Processes View -->
      <div v-if="viewType === 'processes'" class="space-y-2 pr-1">
        <div class="flex items-center gap-2 text-sm font-semibold">
          <span>⚙️</span>
          <span>{{ $t('table.processes') }}</span>
          <span class="text-xs text-net-muted font-normal">({{ processes.length }})</span>
        </div>

        <div v-if="!processes || processes.length === 0" class="text-net-muted text-xs py-8 text-center">{{ $t('table.no_processes') }}</div>

        <div v-else class="space-y-1">
          <div v-for="p in processes" :key="p.pid"
            class="flex items-center gap-2 px-2.5 py-2 bg-net-surface rounded border"
            :class="isBlockedPID(p.pid) ? 'border-red-900/30' : 'border-net-border'"
          >
            <!-- App icon avatar -->
            <span class="w-6 h-6 rounded flex-none flex items-center justify-center text-[10px] font-bold text-white shadow-sm"
              :style="{ backgroundColor: processColor(p.name) }"
            >{{ processInitial(p.name) }}</span>
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="text-xs font-medium truncate">{{ p.name }}</span>
                <span class="font-mono text-[9px] text-net-muted">PID {{ p.pid }}</span>
              </div>
              <div class="flex items-center gap-2 text-[9px] text-net-muted">
                <span>{{ p.connectionCount }} conn</span>
                <span v-if="p.rx_bps_str && p.rx_bps_str !== '—'" class="text-green-400/80">▼{{ p.rx_bps_str }}</span>
                <span v-if="p.tx_bps_str && p.tx_bps_str !== '—'" class="text-blue-400/80">▲{{ p.tx_bps_str }}</span>
                <span v-if="p.estimated" class="text-[8px] opacity-50" title="Estimated from connection share">~</span>
              </div>
            </div>
            <span v-if="isBlockedPID(p.pid)" class="text-[9px] px-1.5 py-0.5 rounded bg-net-danger/10 text-net-danger font-medium">BLOCKED</span>
            <button v-if="!isBlockedPID(p.pid)"
              @click="$emit('block-pid', { pid: p.pid, name: p.name, iface: p.iface })"
              class="text-[9px] px-2 py-1 rounded bg-net-danger/10 text-net-danger border border-net-danger/30 hover:bg-net-danger/20 transition-colors font-medium"
            >{{ $t('table.block') }}</button>
            <button v-else
              @click="unblockPID(p.pid)"
              class="text-[9px] px-2 py-1 rounded bg-net-bg text-net-muted hover:text-net-text border border-net-border hover:border-net-text/30 transition-colors font-medium"
            >{{ $t('table.unblock') }}</button>
          </div>
        </div>
      </div>

      <!-- Interface Detail View -->
      <div v-else-if="!selectedIface" class="flex flex-col items-center justify-center h-full text-net-muted">
        <svg class="w-16 h-16 mb-3 opacity-20" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.858 15.355-5.858 21.213 0" />
        </svg>
        <p class="text-xs opacity-50">{{ $t('table.select_hint') }}</p>
      </div>

      <div v-else class="space-y-3 pr-1">
        <div class="flex items-center gap-3">
          <span class="w-3 h-3 rounded-full flex-none" :class="selectedIface.is_up ? 'bg-net-success' : 'bg-net-danger'"></span>
          <span class="font-mono text-base font-semibold">{{ selectedIface.name }}</span>
          <span class="text-xs px-2 py-0.5 rounded font-medium"
            :class="selectedIface.is_up ? 'bg-net-success/10 text-net-success' : 'bg-net-danger/10 text-net-danger'"
          >{{ selectedIface.is_up ? $t('table.up') : $t('table.down') }}</span>
        </div>

        <div v-if="tp(selectedIface.name)" class="flex items-center gap-4 p-2.5 bg-net-surface rounded border border-net-border">
          <div class="flex items-center gap-1.5">
            <span class="text-[10px] text-net-muted">RX</span>
            <span class="font-mono text-xs text-green-400">{{ tp(selectedIface.name).rx_bps_str }}</span>
          </div>
          <div class="w-px h-4 bg-net-border"></div>
          <div class="flex items-center gap-1.5">
            <span class="text-[10px] text-net-muted">TX</span>
            <span class="font-mono text-xs text-blue-400">{{ tp(selectedIface.name).tx_bps_str }}</span>
          </div>
          <span class="text-[10px] text-net-muted ml-auto">{{ formatBytes(tp(selectedIface.name).rx_bytes) }} / {{ formatBytes(tp(selectedIface.name).tx_bytes) }}</span>
        </div>

        <div class="grid grid-cols-3 gap-3">
          <div class="p-2 bg-net-surface rounded border border-net-border">
            <span class="text-[10px] text-net-muted block">{{ $t('table.mac') }}</span>
            <span class="font-mono text-xs">{{ selectedIface.mac_address || '—' }}</span>
          </div>
          <div class="p-2 bg-net-surface rounded border border-net-border">
            <span class="text-[10px] text-net-muted block">MTU</span>
            <span class="font-mono text-xs">{{ selectedIface.mtu || '—' }}</span>
          </div>
          <div class="p-2 bg-net-surface rounded border border-net-border">
            <span class="text-[10px] text-net-muted block">{{ $t('table.flags') }}</span>
            <div class="flex flex-wrap gap-1 mt-0.5">
              <span v-for="flag in selectedIface.flags" :key="flag" class="text-[9px] px-1 py-0.5 rounded bg-net-bg text-net-muted font-mono">{{ flag }}</span>
              <span v-if="!selectedIface.flags || selectedIface.flags.length === 0" class="text-net-muted text-[10px]">—</span>
            </div>
          </div>
        </div>

        <div class="p-2.5 bg-net-surface rounded border border-net-border">
          <span class="text-[10px] text-net-muted block mb-1.5">{{ $t('table.ip_addresses') }}</span>
          <div v-if="selectedIface.ip_addresses.length === 0" class="text-net-muted text-[10px]">—</div>
          <div v-else class="flex flex-wrap gap-1.5">
            <div v-for="ip in selectedIface.ip_addresses" :key="ip.address"
              class="flex items-center gap-1.5 rounded px-2 py-1 group cursor-context-menu text-[11px]"
              :class="isBlockedIP(ip.address) ? 'bg-red-900/20 border border-red-800/40' : 'bg-net-bg'"
              @contextmenu.prevent="$emit('ctx-ip', $event, { ip: ip.address, iface: selectedIface.name })"
            >
              <span class="text-[9px] font-medium px-1 py-0.5 rounded"
                :class="ip.address_type === 'IPv4' ? 'bg-blue-500/10 text-blue-400' : 'bg-purple-500/10 text-purple-400'"
              >{{ ip.address_type }}</span>
              <span class="font-mono">{{ ip.address }}</span>
              <span v-if="ip.network" class="text-[9px] text-net-muted">/{{ ip.network }}</span>
              <span v-if="isBlockedIP(ip.address)" class="text-[9px] px-1 py-0.5 rounded bg-net-danger/20 text-net-danger font-medium">BLOCKED</span>
              <button @click="copyToClipboard(ip.address)" class="opacity-0 group-hover:opacity-100 text-net-muted hover:text-net-text text-[9px] px-1">{{ $t('table.copy') }}</button>
            </div>
          </div>
        </div>

        <div class="bg-net-surface rounded border border-net-border overflow-hidden">
          <div class="px-2.5 py-1.5 border-b border-net-border flex items-center justify-between">
            <span class="text-[10px] text-net-muted">
              {{ $t('table.connections') }}
              <span v-if="selectedIface.connections.length > 0">({{ selectedIface.connections.length }})</span>
            </span>
          </div>
          <div v-if="selectedIface.connections.length === 0" class="p-3 text-net-muted text-[10px] text-center">{{ $t('table.no_connections') }}</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-[10px] border-collapse">
              <thead>
                <tr class="border-b border-net-border text-net-muted">
                  <th class="text-left py-1 px-1 font-medium cursor-pointer hover:text-net-text" @click="setSort('protocol')">{{ $t('table.protocol') }}<span class="text-[9px]" v-html="sortArrow('protocol')"></span></th>
                  <th class="text-left py-1 px-1 font-medium cursor-pointer hover:text-net-text" @click="setSort('local_port')">{{ $t('table.local_port') }}<span class="text-[9px]" v-html="sortArrow('local_port')"></span></th>
                  <th class="text-left py-1 px-1 font-medium cursor-pointer hover:text-net-text" @click="setSort('remote_ip')">{{ $t('table.remote') }}<span class="text-[9px]" v-html="sortArrow('remote_ip')"></span></th>
                  <th class="text-left py-1 px-1 font-medium cursor-pointer hover:text-net-text" @click="setSort('status')">{{ $t('table.state') }}<span class="text-[9px]" v-html="sortArrow('status')"></span></th>
                  <th class="text-left py-1 px-1 font-medium cursor-pointer hover:text-net-text" @click="setSort('pid')">{{ $t('table.pid') }}<span class="text-[9px]" v-html="sortArrow('pid')"></span></th>
                  <th class="text-left py-1 font-medium cursor-pointer hover:text-net-text" @click="setSort('process_name')">{{ $t('table.process') }}<span class="text-[9px]" v-html="sortArrow('process_name')"></span></th>
                  <th class="text-left py-1 px-1 font-medium text-net-muted">▼ RX</th>
                  <th class="text-left py-1 px-1 font-medium text-net-muted">▲ TX</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(conn, idx) in sortedConns(selectedIface.connections)" :key="idx"
                  class="border-b border-net-border/30 transition-colors cursor-context-menu"
                  :class="isConnBlocked(conn) ? 'bg-red-900/10 hover:bg-red-900/20' : 'hover:bg-net-hover/50'"
                  @contextmenu.prevent="$emit('ctx-conn', $event, {
                    protocol: conn.protocol, local_port: conn.local_port,
                    remote_ip: conn.remote_ip, remote_port: conn.remote_port,
                    status: conn.status, pid: conn.pid, process_name: conn.process_name, iface: selectedIface.name
                  })"
                >
                  <td class="py-1 px-1">
                    <span class="font-mono px-1 py-0.5 rounded text-[9px]"
                      :class="conn.protocol === 'TCP' ? 'bg-green-500/10 text-green-400' : 'bg-orange-500/10 text-orange-400'"
                    >{{ conn.protocol }}</span>
                    <span v-if="isConnBlocked(conn)" class="ml-0.5 text-net-danger">⛔</span>
                  </td>
                  <td class="py-1 px-1 font-mono">{{ conn.local_port || '—' }}</td>
                  <td class="py-1 px-1 font-mono text-net-muted">{{ formatRemote(conn) }}</td>
                  <td class="py-1 px-1"><span class="font-mono" :class="statusClass(conn.status)">{{ conn.status || '—' }}</span></td>
                  <td class="py-1 px-1 font-mono text-net-muted">{{ conn.pid > 0 ? conn.pid : '—' }}</td>
                  <td class="py-1 font-mono text-net-muted max-w-[120px] truncate" :title="conn.process_name">
                    <span v-if="conn.pid > 0 && conn.process_name" class="flex items-center gap-1">
                      <span class="w-3.5 h-3.5 rounded flex-none flex items-center justify-center text-[7px] font-bold text-white shadow-sm"
                        :style="{ backgroundColor: processColor(conn.process_name) }"
                      >{{ processInitial(conn.process_name) }}</span>
                      <span class="truncate">{{ conn.process_name }}</span>
                    </span>
                    <span v-else>—</span>
                  </td>
                  <td class="py-1 px-1 font-mono text-[9px]" :class="pidTpMap[conn.pid] && pidTpMap[conn.pid].rx !== '—' ? 'text-green-400/80' : 'text-net-muted/30'">{{ pidTpMap[conn.pid] ? pidTpMap[conn.pid].rx : '—' }}<span v-if="pidTpMap[conn.pid] && !pidTpMap[conn.pid].real && pidTpMap[conn.pid].rx !== '—'" class="text-[7px] opacity-40 ml-0.5">~</span></td>
                  <td class="py-1 px-1 font-mono text-[9px]" :class="pidTpMap[conn.pid] && pidTpMap[conn.pid].tx !== '—' ? 'text-blue-400/80' : 'text-net-muted/30'">{{ pidTpMap[conn.pid] ? pidTpMap[conn.pid].tx : '—' }}<span v-if="pidTpMap[conn.pid] && !pidTpMap[conn.pid].real && pidTpMap[conn.pid].tx !== '—'" class="text-[7px] opacity-40 ml-0.5">~</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  interfaces: Array,
  loading: Boolean,
  locale: String,
  blockedItems: { type: Array, default: () => [] },
  throughput: { type: Array, default: () => [] },
  processes: { type: Array, default: () => [] },
  pidIOData: { type: Array, default: () => [] },
})

const emit = defineEmits(['ctx-ip', 'ctx-conn', 'unblock', 'block-pid', 'unblock-pid'])

// View state: 'iface' or 'blocked'
const viewType = ref('iface')
const viewId = ref(null)
const sortKey = ref('local_port')
const sortDir = ref('asc')

// Computed selections
const selectedIface = computed(() => {
  if (viewType.value !== 'iface' || !viewId.value || !props.interfaces) return null
  return props.interfaces.find(i => i.name === viewId.value) || null
})

// PID throughput lookup (real ETW data preferred, proportional fallback from processes)
const pidTpMap = computed(() => {
  const m = {}
  // Real ETW data
  for (const io of props.pidIOData || []) {
    if (io.pid > 0 && (io.bytes_recv > 0 || io.bytes_sent > 0)) {
      m[io.pid] = { rx: formatBits(io.bytes_recv * 8), tx: formatBits(io.bytes_sent * 8), real: true }
    }
  }
  // Fill in missing PIDs with proportional estimates
  for (const p of props.processes || []) {
    if (!m[p.pid]) {
      m[p.pid] = { rx: p.rx_bps_str, tx: p.tx_bps_str, real: false }
    }
  }
  return m
})

// Blocked items computed
const blockedIPs = computed(() =>
  props.blockedItems.filter(b => b.type === 'ip' && b.active !== false)
)
const blockedPIDs = computed(() =>
  props.blockedItems.filter(b => b.type === 'pid' && b.active !== false)
)
const blockedCount = computed(() => blockedIPs.value.length + blockedPIDs.value.length)

// Select view
function selectView(type, id) {
  viewType.value = type
  viewId.value = id || null
}

// Auto-select first interface
watch(() => props.interfaces, (val) => {
  if (val && val.length > 0 && viewType.value === 'iface' && !viewId.value) {
    viewId.value = val[0].name
  }
}, { immediate: true })

// Blocked helpers
function isBlockedIP(ip) {
  return props.blockedItems.some(b => b.type === 'ip' && b.target === ip && b.active !== false)
}
function unblockPID(pid) {
  const entry = props.blockedItems.find(b => b.type === 'pid' && b.target.includes(`PID ${pid}`) && b.active !== false)
  if (entry) emit('unblock', entry.id)
}
function isBlockedPID(pid) {
  return props.blockedItems.some(b => b.type === 'pid' && b.target.includes(`PID ${pid}`) && b.active !== false)
}
function isConnBlocked(conn) {
  if (conn.remote_ip && conn.remote_ip !== '0.0.0.0' && conn.remote_ip !== '::' && isBlockedIP(conn.remote_ip)) return true
  if (conn.pid > 0 && isBlockedPID(conn.pid)) return true
  return false
}

// Throughput
function tp(ifaceName) { return props.throughput.find(t => t.interface === ifaceName) }
function formatBytes(bytes) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0; let val = bytes
  while (val >= 1024 && i < units.length - 1) { val /= 1024; i++ }
  return val.toFixed(1) + ' ' + units[i]
}

// Sorting
function sortArrow(key) { return sortKey.value === key ? (sortDir.value === 'asc' ? '▲' : '▼') : '' }
function setSort(key) {
  if (sortKey.value === key) { sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc' }
  else { sortKey.value = key; sortDir.value = 'asc' }
}
function sortedConns(connections) {
  if (!connections) return []
  const sorted = [...connections]; const key = sortKey.value; const dir = sortDir.value === 'asc' ? 1 : -1
  sorted.sort((a, b) => {
    let va = a[key], vb = b[key]
    if (typeof va === 'string') va = va.toLowerCase()
    if (typeof vb === 'string') vb = vb.toLowerCase()
    if (va == null) va = ''; if (vb == null) vb = ''
    return va < vb ? -1 * dir : va > vb ? 1 * dir : 0
  })
  return sorted
}

function formatRemote(conn) {
  if (conn.status === 'LISTEN') return '*:*'
  if (conn.remote_ip && conn.remote_ip !== '0.0.0.0' && conn.remote_ip !== '::') return conn.remote_ip + ':' + conn.remote_port
  return '*:*'
}
function statusClass(status) {
  if (!status) return 'text-net-muted'
  const s = status.toUpperCase()
  if (s === 'LISTEN') return 'text-green-400'
  if (s === 'ESTABLISHED') return 'text-net-accent'
  if (s === 'TIME_WAIT' || s === 'CLOSE_WAIT') return 'text-yellow-400'
  return 'text-net-muted'
}
// Process icon helpers
const avatarColors = [
  '#3B82F6', '#EF4444', '#10B981', '#F59E0B', '#8B5CF6',
  '#EC4899', '#06B6D4', '#84CC16', '#F97316', '#6366F1',
  '#14B8A6', '#D946EF', '#0EA5E9', '#22C55E', '#EAB308',
]
function processColor(name) {
  if (!name) return avatarColors[0]
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = ((hash << 5) - hash) + name.charCodeAt(i)
  return avatarColors[Math.abs(hash) % avatarColors.length]
}
function processInitial(name) {
  return (name && name.length > 0) ? name[0].toUpperCase() : '?'
}

function copyToClipboard(text) {
  try { navigator.clipboard.writeText(text) } catch {
    const ta = document.createElement('textarea'); ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0'
    document.body.appendChild(ta); ta.select(); document.execCommand('copy'); document.body.removeChild(ta)
  }
}
</script>
