<template>
  <div :dir="currentDir" class="h-screen w-screen flex flex-col overflow-hidden select-none" @contextmenu.prevent @click="closeContextMenu">
    <header class="flex-none px-4 py-2 bg-net-surface border-b border-net-border flex items-center gap-3">
      <svg class="w-5 h-5 text-net-accent flex-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
      </svg>
      <h1 class="text-sm font-semibold text-net-text tracking-wide">{{ $t('app.title') }}</h1>
      <span class="text-xs text-net-muted hidden sm:inline">v1.0</span>
    </header>

    <Toolbar :scanning="scanning" :locale="currentLocale"
      @scan="triggerScan"
      @toggle-lang="toggleLanguage"
      @update:filter="filterText = $event"
    />

    <main class="flex-1 overflow-auto p-3">
      <NetworkTable
        :interfaces="filteredInterfaces"
        :loading="scanning"
        :locale="currentLocale"
        :blocked-items="blockedItems"
        :throughput="throughputData"
        :pid-io-data="pidIOData"
        @ctx-ip="onContextIP"
        @ctx-conn="onContextConn"
        @unblock="onUnblockFromList"
        @block-pid="onBlockPID"
        :processes="processesList"
      />
    </main>

    <StatusBar :status="statusMessage" :has-permission-error="hasPermissionError" />

    <ContextMenu :visible="ctxVisible" :x="ctxX" :y="ctxY" :items="ctxItems"
      @close="closeContextMenu" @action="onContextAction" />
    <ConfirmDialog :visible="confirmVisible" :title="confirmTitle" :message="confirmMessage"
      :confirm-text="confirmBtn" :cancel-text="t('table.cancel')" :danger="confirmDanger"
      @confirm="onConfirm" @cancel="confirmVisible = false" />
    <ToastNotification :toasts="toasts" :locale="currentLocale" @dismiss="dismissToast" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { QuickScan, BlockTarget, UnblockTarget, GetThroughput, GetProcessNetIO } from '../wailsjs/go/main/App'
import Toolbar from './components/Toolbar.vue'
import NetworkTable from './components/NetworkTable.vue'
import StatusBar from './components/StatusBar.vue'
import ContextMenu from './components/ContextMenu.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import ToastNotification from './components/ToastNotification.vue'

const { t, locale } = useI18n()

const scanning = ref(false)
const scanResult = ref(null)
const filterText = ref('')
const statusMessage = ref('')
const hasPermissionError = ref(false)
const blockedItems = ref([])
const throughputData = ref([])
const pidIOData = ref([])
let throughputTimer = null

const currentLocale = computed(() => locale.value)
const currentDir = computed(() => locale.value === 'ar' ? 'rtl' : 'ltr')

const processesList = computed(() => {
  if (!scanResult.value || !scanResult.value.interfaces) return []
  const pidMap = new Map()
  const ifaceConnCounts = {}
  const ifacePidCounts = {}

  // Build ETW lookup map for real per-PID counters
  const etwMap = {}
  for (const io of pidIOData.value) {
    if (io.pid > 0) {
      etwMap[io.pid] = { rx: io.bytes_recv, tx: io.bytes_sent }
    }
  }

  // Count connections per (PID, interface) and per interface
  for (const iface of scanResult.value.interfaces) {
    const ifaceName = iface.name
    if (!ifaceConnCounts[ifaceName]) ifaceConnCounts[ifaceName] = 0
    for (const conn of iface.connections || []) {
      if (conn.pid <= 0) continue
      ifaceConnCounts[ifaceName]++
      const key = conn.pid + '|' + ifaceName
      ifacePidCounts[key] = (ifacePidCounts[key] || 0) + 1

      if (!pidMap.has(conn.pid)) {
        const etw = etwMap[conn.pid]
        const hasReal = etw && (etw.rx > 0 || etw.tx > 0)
        pidMap.set(conn.pid, {
          pid: conn.pid,
          name: conn.process_name || 'Unknown',
          iface: ifaceName,
          connectionCount: 0,
          rx_bps: 0,
          tx_bps: 0,
          rx_bps_str: hasReal ? formatBits(etw.rx * 8) : '—',
          tx_bps_str: hasReal ? formatBits(etw.tx * 8) : '—',
          estimated: !hasReal,
        })
      }
      pidMap.get(conn.pid).connectionCount++
    }
  }

  // Allocate bandwidth proportionally from throughput data (fallback for PIDs without ETW)
  for (const td of throughputData.value) {
    const ifaceName = td.interface
    const totalConns = ifaceConnCounts[ifaceName]
    if (!totalConns || totalConns === 0) continue
    for (const [pid, proc] of pidMap) {
      if (!proc.estimated) continue // skip if real ETW data exists
      const key = pid + '|' + ifaceName
      const pidConns = ifacePidCounts[key] || 0
      if (pidConns === 0) continue
      const share = pidConns / totalConns
      proc.rx_bps += td.rx_bps * share
      proc.tx_bps += td.tx_bps * share
    }
  }

  // Format bit rates for estimated PIDs
  for (const proc of pidMap.values()) {
    if (proc.estimated) {
      proc.rx_bps_str = formatBits(proc.rx_bps)
      proc.tx_bps_str = formatBits(proc.tx_bps)
    }
  }

  return Array.from(pidMap.values()).sort((a, b) => a.pid - b.pid)
})

const filteredInterfaces = computed(() => {
  if (!scanResult.value || !scanResult.value.interfaces) return []
  const filter = filterText.value.toLowerCase()
  if (!filter) return scanResult.value.interfaces
  return scanResult.value.interfaces.filter(iface =>
    iface.name.toLowerCase().includes(filter)
  )
})

// ---- Throughput Polling ----
function startThroughputPolling() {
  stopThroughputPolling()
  throughputTimer = setInterval(async () => {
    try {
      throughputData.value = await GetThroughput()
    } catch { /* ignore polling errors */ }

    // Also poll per-process network IO data (ETW on Windows)
    try {
      pidIOData.value = await GetProcessNetIO()
    } catch { /* ignore */ }
  }, 1000)
}

function stopThroughputPolling() {
  if (throughputTimer) {
    clearInterval(throughputTimer)
    throughputTimer = null
  }
}

// ---- Context Menu State ----
const ctxVisible = ref(false)
const ctxX = ref(0)
const ctxY = ref(0)
const ctxItems = ref([])
let ctxData = null

function closeContextMenu() { ctxVisible.value = false }

function buildCtxMenu(items) { ctxItems.value = items }

function isBlockedIP(ip) {
  return blockedItems.value.some(b =>
    (b.type === 'ip' && b.target === ip && b.active !== false)
  )
}

function isBlockedPID(pid) {
  return blockedItems.value.some(b =>
    b.type === 'pid' && b.target.includes(`PID ${pid}`) && b.active !== false
  )
}

function getBlockIDForIP(ip) {
  const entry = blockedItems.value.find(b => b.type === 'ip' && b.target === ip)
  return entry ? entry.id : null
}

function onContextIP(event, data) {
  ctxData = { type: 'ip', target: data.ip, iface: data.iface }
  ctxX.value = event.clientX
  ctxY.value = event.clientY

  const blocked = isBlockedIP(data.ip)
  buildCtxMenu(
    blocked
      ? [{ label: '🔓 ' + t('table.unblock') + ' ' + data.ip, action: 'unblock-ip' }]
      : [{ label: '🔒 ' + t('table.block_ip') + ' ' + data.ip, action: 'block-ip', danger: true }]
  )
  ctxVisible.value = true
}

function onContextConn(event, data) {
  ctxData = { type: 'conn', ...data }
  ctxX.value = event.clientX
  ctxY.value = event.clientY

  const items = []
  if (data.remote_ip && data.remote_ip !== '0.0.0.0' && data.remote_ip !== '::') {
    const blocked = isBlockedIP(data.remote_ip)
    items.push(blocked
      ? { label: '🔓 ' + t('table.unblock') + ' ' + data.remote_ip, action: 'unblock-conn-ip' }
      : { label: '🔒 ' + t('table.block_ip') + ' ' + data.remote_ip, action: 'block-conn-ip', danger: true }
    )
  }
  if (data.pid > 0) {
    const blocked = isBlockedPID(data.pid)
    items.push(blocked
      ? { label: '🔓 ' + t('table.unblock') + ' PID ' + data.pid, action: 'unblock-pid' }
      : { label: '🔒 ' + t('table.block_pid') + ' ' + data.pid, action: 'block-pid', danger: true }
    )
  }
  if (items.length === 0) {
    items.push({ label: '— ' + t('table.no_connections'), action: '' })
  }
  buildCtxMenu(items)
  ctxVisible.value = true
}

// ---- Confirm Dialog State ----
const confirmVisible = ref(false)
const confirmTitle = ref('')
const confirmMessage = ref('')
const confirmBtn = ref('')
const confirmDanger = ref(false)
let pendingAction = null

function showConfirm(title, message, btn, danger, action) {
  confirmTitle.value = title
  confirmMessage.value = message
  confirmBtn.value = btn
  confirmDanger.value = danger
  pendingAction = action
  confirmVisible.value = true
}

function onConfirm() {
  confirmVisible.value = false
  if (pendingAction) pendingAction()
  pendingAction = null
}

// ---- Toast State ----
const toasts = ref([])
let toastId = 0

function addToast(type, message, duration = 4000) {
  const id = ++toastId
  toasts.value = [...toasts.value, { id, type, message }]
  setTimeout(() => dismissToast(id), duration)
}

function dismissToast(id) {
  toasts.value = toasts.value.filter(t => t.id !== id)
}

// ---- Blocking Actions ----
async function executeBlock(type, target, ifaceName) {
  try {
    const result = await BlockTarget({ type, target, interface: ifaceName || '' })
    if (result.success) {
      addToast('success', result.message)
      if (result.entry) {
        blockedItems.value = [...blockedItems.value, result.entry]
      }
    } else {
      addToast('error', result.message)
    }
  } catch (err) {
    addToast('error', t('status.error', { error: err.message }))
  }
}

async function executeBlockPID(pid) {
  try {
    const result = await BlockTarget({ type: 'pid', target: String(pid), pid: parseInt(pid) })
    if (result.success) {
      addToast('success', result.message)
    } else {
      addToast('error', result.message)
    }
  } catch (err) {
    addToast('error', t('status.error', { error: err.message }))
  }
}

async function executeUnblock(blockId) {
  try {
    const result = await UnblockTarget(blockId)
    if (result.success) {
      addToast('success', result.message)
      blockedItems.value = blockedItems.value.filter(b => b.id !== blockId)
    } else {
      addToast('error', result.message)
    }
  } catch (err) {
    addToast('error', t('status.error', { error: err.message }))
  }
}

function onContextAction(action) {
  if (!ctxData) return

  switch (action) {
    case 'block-ip':
      showConfirm(t('table.block_ip'), `Block IP ${ctxData.target}?`, t('table.block'), true,
        () => executeBlock('ip', ctxData.target, ctxData.iface))
      break

    case 'block-conn-ip':
      showConfirm(t('table.block_ip'), `Block remote IP ${ctxData.remote_ip}?`, t('table.block'), true,
        () => executeBlock('ip', ctxData.remote_ip, ctxData.iface))
      break

    case 'block-pid':
      showConfirm(t('table.block_pid'), `Block PID ${ctxData.pid} (${ctxData.process_name || 'unknown'})?`,
        t('table.block'), true, () => executeBlockPID(ctxData.pid))
      break

    case 'unblock-ip':
      showConfirm(t('table.unblock'), `Unblock IP ${ctxData.target}?`, t('table.unblock'), false,
        () => executeUnblock('ip_' + ctxData.target))
      break

    case 'unblock-conn-ip':
      showConfirm(t('table.unblock'), `Unblock IP ${ctxData.remote_ip}?`, t('table.unblock'), false,
        () => executeUnblock('ip_' + ctxData.remote_ip))
      break

    case 'unblock-pid':
      showConfirm(t('table.unblock'), `Unblock PID ${ctxData.pid}?`, t('table.unblock'), false,
        () => {
          blockedItems.value.filter(b => b.type === 'pid' && b.target.includes(`PID ${ctxData.pid}`))
            .forEach(b => executeUnblock(b.id))
        })
      break
  }
}

function onUnblockFromList(blockId) {
  showConfirm(t('table.unblock'), t('table.unblock_confirm'), t('table.unblock'), false,
    () => executeUnblock(blockId))
}

function onBlockPID(data) {
  showConfirm(t('table.block_pid'), `Block PID ${data.pid} (${data.name || 'unknown'})?`, t('table.block'), true,
    () => executeBlockPID(data.pid))
}

function onUnblockPID(blockId) {
  showConfirm(t('table.unblock'), t('table.unblock_confirm'), t('table.unblock'), false,
    () => executeUnblock(blockId))
}

// ---- Scan ----
async function triggerScan() {
  scanning.value = true
  hasPermissionError.value = false
  statusMessage.value = t('status.scanning')
  startThroughputPolling()

  try {
    const result = await QuickScan()
    scanResult.value = result

    if (result.error) {
      if (result.permission_error) {
        hasPermissionError.value = true
        statusMessage.value = t('status.permission_warning')
        addToast('warning', t('status.permission_warning'), 6000)
      } else {
        statusMessage.value = t('status.error', { error: result.error })
        addToast('error', statusMessage.value)
      }
    } else {
      statusMessage.value = t('status.completed', {
        duration: result.duration_ms, interfaces: result.total_interfaces, connections: result.total_connections
      })
      addToast('success', statusMessage.value, 3000)
    }
  } catch (err) {
    statusMessage.value = t('status.error', { error: err.message || 'Unknown error' })
    addToast('error', statusMessage.value)
  } finally {
    scanning.value = false
  }
}

function toggleLanguage() {
  locale.value = locale.value === 'en' ? 'ar' : 'en'
  document.documentElement.lang = locale.value
  document.documentElement.dir = locale.value === 'ar' ? 'rtl' : 'ltr'
}

function formatBits(bps) {
  if (!bps || bps <= 0) return '—'
  if (bps >= 1_000_000_000) return (bps / 1_000_000_000).toFixed(2) + ' Gbps'
  if (bps >= 1_000_000) return (bps / 1_000_000).toFixed(2) + ' Mbps'
  if (bps >= 1_000) return (bps / 1_000).toFixed(2) + ' Kbps'
  return Math.round(bps) + ' bps'
}

onMounted(() => {
  statusMessage.value = t('status.ready')
  document.documentElement.lang = locale.value
  document.documentElement.dir = locale.value === 'ar' ? 'rtl' : 'ltr'
})

onUnmounted(() => {
  stopThroughputPolling()
})
</script>
