<template>
  <el-dialog
    v-model="visible"
    :title="`终端 — ${podName}`"
    width="860px"
    top="4vh"
    @open="onOpen"
    @close="onClose"
    :close-on-click-modal="false"
  >
    <!-- 工具栏 -->
    <div class="term-toolbar">
      <el-tag type="info" size="small">容器</el-tag>

      <!-- 多容器选择 -->
      <el-select
        v-if="containers.length > 1"
        v-model="currentContainer"
        size="small"
        style="width:150px"
        @change="reconnect"
      >
        <el-option v-for="c in containers" :key="c" :label="c" :value="c" />
      </el-select>
      <el-tag v-else size="small" effect="plain">{{ currentContainer || '-' }}</el-tag>

      <el-divider direction="vertical" />

      <!-- 连接状态 -->
      <el-tag :type="wsStatus.type" size="small" effect="light">
        {{ wsStatus.text }}
      </el-tag>

      <el-button size="small" :icon="Refresh" @click="reconnect">重连</el-button>
      <el-button size="small" :icon="Delete"  @click="clearTerminal">清屏</el-button>
    </div>

    <!-- 终端容器 -->
    <div ref="terminalEl" class="terminal-wrap" />

    <template #footer>
      <span class="term-tip">点击终端后即可输入命令</span>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { WebLinksAddon } from 'xterm-addon-web-links'
import 'xterm/css/xterm.css'
import { Refresh, Delete } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: Boolean,
  namespace:  String,
  podName:    String,
  containers: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const visible          = ref(false)
const terminalEl       = ref(null)
const currentContainer = ref('')
let   term             = null   // xterm 实例
let   fitAddon         = null
let   ws               = null   // WebSocket 实例
let   resizeObserver   = null

// 连接状态
const status = ref('disconnected') // connecting | connected | disconnected
const wsStatus = computed(() => ({
  connecting:   { type: 'warning', text: '连接中...' },
  connected:    { type: 'success', text: '已连接' },
  disconnected: { type: 'info',    text: '未连接' },
  error:        { type: 'danger',  text: '连接失败' },
}[status.value]))

watch(() => props.modelValue, val => { visible.value = val })
watch(visible, val => emit('update:modelValue', val))

// ── 初始化 xterm ──────────────────────────────────────
function initTerminal() {
  // 销毁旧实例
  if (term) { term.dispose(); term = null }

  term = new Terminal({
    cursorBlink:    true,
    fontSize:       14,
    fontFamily:     "'Consolas','Monaco','Courier New',monospace",
    theme: {
      background: '#1a1a2e',
      foreground: '#d4d4d4',
      cursor:     '#ffffff',
      black:      '#1a1a2e',
      green:      '#67c23a',
      yellow:     '#e6a23c',
      red:        '#f56c6c',
      cyan:       '#409eff',
    },
    rows: 30,
    cols: 100,
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.loadAddon(new WebLinksAddon())
  term.open(terminalEl.value)
  fitAddon.fit()

  // 用户在终端里输入 → 通过 WebSocket 发给后端
  term.onData(data => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'input', data }))
    }
  })

  // 监听终端大小变化 → 告诉后端调整 Pod 终端尺寸
  resizeObserver = new ResizeObserver(() => {
    fitAddon?.fit()
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'resize',
        rows: term.rows,
        cols: term.cols,
      }))
    }
  })
  resizeObserver.observe(terminalEl.value)
}

// ── 建立 WebSocket 连接 ───────────────────────────────
function connect() {
  if (ws) { ws.close(); ws = null }

  const container = currentContainer.value || props.containers[0] || ''
  const protocol  = location.protocol === 'https:' ? 'wss' : 'ws'
  // vite dev proxy 会把 /api 转发到后端，WebSocket 也一样
  const url = `${protocol}://${location.host}/api/pods/${props.namespace}/${props.podName}/exec?container=${container}`

  status.value = 'connecting'
  ws = new WebSocket(url)

  ws.onopen = () => {
    status.value = 'connected'
    // 连上后立即同步终端大小
    ws.send(JSON.stringify({ type: 'resize', rows: term.rows, cols: term.cols }))
    term.focus()
  }

  // 后端发来的是 Pod 输出，直接写入 xterm
  ws.onmessage = (e) => {
    term?.write(e.data)
  }

  ws.onerror = () => { status.value = 'error' }

  ws.onclose = () => {
    if (status.value === 'connected') {
      term?.write('\r\n\x1b[33m[连接已断开]\x1b[0m\r\n')
    }
    status.value = 'disconnected'
  }
}

function reconnect() {
  term?.clear()
  connect()
}

function clearTerminal() {
  term?.clear()
}

// ── 弹窗开启/关闭 ─────────────────────────────────────
async function onOpen() {
  currentContainer.value = props.containers[0] ?? ''
  await nextTick()
  initTerminal()
  connect()
}

function onClose() {
  ws?.close()
  ws = null
  resizeObserver?.disconnect()
  resizeObserver = null
  term?.dispose()
  term = null
  status.value = 'disconnected'
}

onUnmounted(onClose)
</script>

<style scoped>
.term-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}
.terminal-wrap {
  background: #1a1a2e;
  border-radius: 6px;
  padding: 8px;
  height: 460px;
  overflow: hidden;
}
.term-tip {
  font-size: 12px;
  color: #999;
  margin-right: auto;
}
</style>
