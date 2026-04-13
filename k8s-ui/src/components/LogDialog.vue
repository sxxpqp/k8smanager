<template>
  <el-dialog
    v-model="visible"
    :title="`日志 — ${podName}`"
    width="960px"
    top="3vh"
    @open="onOpen"
    @close="onClose"
  >
    <!-- 工具栏 -->
    <div class="log-toolbar">
      <!-- 容器选择（多容器时显示）-->
      <template v-if="containers.length > 1">
        <el-tag type="info" size="small">容器</el-tag>
        <el-select
          v-model="currentContainer"
          size="small"
          style="width:160px"
          @change="fetchLogs"
        >
          <el-option
            v-for="c in containers"
            :key="c"
            :label="c"
            :value="c"
          />
        </el-select>
        <el-divider direction="vertical" />
      </template>

      <!-- 行数 -->
      <el-select v-model="tail" @change="fetchLogs" style="width:120px" size="small">
        <el-option label="50 行"  :value="50" />
        <el-option label="100 行" :value="100" />
        <el-option label="200 行" :value="200" />
        <el-option label="500 行" :value="500" />
        <el-option label="1000 行" :value="1000" />
      </el-select>

      <!-- 关键字搜索 -->
      <el-input
        v-model="keyword"
        placeholder="关键字高亮..."
        clearable
        size="small"
        style="width:160px"
        :prefix-icon="Search"
      />

      <el-switch
        v-model="autoRefresh"
        active-text="自动刷新"
        inactive-text=""
        size="small"
        @change="toggleAutoRefresh"
      />

      <el-button :icon="Refresh" size="small" @click="fetchLogs" :loading="loading" />
      <el-button :icon="CopyDocument" size="small" @click="copyLogs" />
      <el-button :icon="Bottom" size="small" @click="scrollToBottom" />

      <!-- 行数统计 -->
      <span class="log-stat">{{ lineCount }} 行</span>
    </div>

    <!-- 日志内容 -->
    <div class="log-container" ref="logContainer">
      <div v-if="loading && !logText" class="log-loading">
        <el-icon class="is-loading" :size="20"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      <pre v-else class="log-content" v-html="highlightedLog || '暂无日志'" />
    </div>

    <template #footer>
      <span class="log-footer-info">
        容器: {{ currentContainer || containers[0] || '-' }}
      </span>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, nextTick, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, CopyDocument, Bottom, Loading, Search } from '@element-plus/icons-vue'
import { getPodLogs } from '../api/index.js'

const props = defineProps({
  modelValue:  Boolean,
  namespace:   String,
  podName:     String,
  // 容器列表，由父组件传入，如 ["nginx", "sidecar"]
  containers:  { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue'])

const visible          = ref(false)
const logText          = ref('')
const loading          = ref(false)
const tail             = ref(200)
const keyword          = ref('')
const autoRefresh      = ref(false)
const currentContainer = ref('')
const logContainer     = ref(null)
let   timer            = null

// 同步 v-model
watch(() => props.modelValue, val => { visible.value = val })
watch(visible, val => emit('update:modelValue', val))

// 切换 Pod 时重置容器选择
watch(() => props.podName, () => {
  currentContainer.value = props.containers[0] ?? ''
  logText.value = ''
})

// 日志行数
const lineCount = computed(() => logText.value ? logText.value.split('\n').length : 0)

// 关键字高亮
const highlightedLog = computed(() => {
  if (!logText.value) return ''
  const escaped = logText.value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  if (!keyword.value.trim()) return escaped
  const kw = keyword.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return escaped.replace(
    new RegExp(kw, 'gi'),
    m => `<mark class="log-highlight">${m}</mark>`
  )
})

async function fetchLogs() {
  if (!props.namespace || !props.podName) return
  loading.value = true
  try {
    const container = currentContainer.value || props.containers[0] || ''
    const res = await getPodLogs(props.namespace, props.podName, tail.value, container)
    logText.value = typeof res === 'string' ? res : (res.data ?? '')
    await nextTick()
    scrollToBottom()
  } catch (e) {
    ElMessage.error('获取日志失败: ' + e.message)
  } finally {
    loading.value = false
  }
}

function onOpen() {
  currentContainer.value = props.containers[0] ?? ''
  fetchLogs()
}

function onClose() {
  stopAutoRefresh()
  logText.value = ''
  keyword.value = ''
}

function scrollToBottom() {
  if (logContainer.value)
    logContainer.value.scrollTop = logContainer.value.scrollHeight
}

function copyLogs() {
  navigator.clipboard.writeText(logText.value)
    .then(() => ElMessage.success('已复制到剪贴板'))
    .catch(() => ElMessage.error('复制失败'))
}

function toggleAutoRefresh(val) {
  clearInterval(timer)
  if (val) timer = setInterval(fetchLogs, 5000)
}

function stopAutoRefresh() {
  autoRefresh.value = false
  clearInterval(timer)
  timer = null
}

onUnmounted(stopAutoRefresh)
</script>

<style scoped>
.log-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}
.log-stat {
  margin-left: auto;
  font-size: 12px;
  color: #666;
}
.log-container {
  background: #1a1a2e;
  border-radius: 6px;
  height: 540px;
  overflow-y: auto;
  padding: 12px 16px;
}
.log-loading {
  color: #888;
  display: flex;
  align-items: center;
  gap: 8px;
  justify-content: center;
  height: 100%;
}
.log-content {
  color: #d4d4d4;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}
.log-footer-info {
  font-size: 12px;
  color: #999;
  margin-right: auto;
}
</style>

<!-- 全局：高亮样式不能 scoped -->
<style>
.log-highlight {
  background: #f0c040;
  color: #000;
  border-radius: 2px;
  padding: 0 1px;
}
</style>
