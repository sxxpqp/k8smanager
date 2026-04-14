<template>
  <el-dialog
    v-model="visible"
    :title="`ConfigMap — ${cmName}`"
    width="780px"
    top="4vh"
    @open="fetchDetail"
  >
    <div v-loading="loading" style="min-height:200px">
      <!-- Key 列表 -->
      <div class="cm-layout">
        <!-- 左：key 列表 -->
        <div class="cm-keys">
          <div class="cm-keys-title">键列表</div>
          <div
            v-for="key in keys"
            :key="key"
            class="cm-key-item"
            :class="{ active: activeKey === key }"
            @click="activeKey = key"
          >
            <el-icon size="12"><Document /></el-icon>
            <span>{{ key }}</span>
          </div>
          <!-- 新增 key -->
          <div class="cm-add-key" v-if="!adding">
            <el-button link :icon="Plus" size="small" @click="adding=true">新增键</el-button>
          </div>
          <div v-else class="cm-new-key">
            <el-input v-model="newKey" size="small" placeholder="键名" autofocus
              @keyup.enter="addKey" @keyup.esc="adding=false" />
            <el-button size="small" type="primary" @click="addKey">确定</el-button>
          </div>
        </div>

        <!-- 右：值编辑 -->
        <div class="cm-value">
          <div v-if="!activeKey" class="cm-empty">← 选择左侧 Key 查看或编辑</div>
          <template v-else>
            <div class="cm-value-header">
              <span class="cm-value-key">{{ activeKey }}</span>
              <el-button link type="danger" :icon="Delete" size="small"
                @click="removeKey(activeKey)">删除此键</el-button>
            </div>
            <el-input
              v-model="editData[activeKey]"
              type="textarea"
              :rows="18"
              style="font-family:monospace; font-size:13px"
              placeholder="键值内容..."
            />
          </template>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存更新</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, Plus, Delete } from '@element-plus/icons-vue'
import { getConfigMapDetail, updateConfigMap } from '../api/index.js'

const props = defineProps({
  modelValue: Boolean,
  namespace:  String,
  cmName:     String
})
const emit = defineEmits(['update:modelValue', 'updated'])

const visible  = ref(false)
const loading  = ref(false)
const saving   = ref(false)
const editData = ref({})   // { key: value }
const activeKey = ref('')
const adding   = ref(false)
const newKey   = ref('')

watch(() => props.modelValue, v => { visible.value = v })
watch(visible, v => emit('update:modelValue', v))

const keys = computed(() => Object.keys(editData.value))

async function fetchDetail() {
  loading.value  = true
  activeKey.value = ''
  editData.value  = {}
  try {
    const res = await getConfigMapDetail(props.namespace, props.cmName)
    editData.value  = { ...(res.data?.data ?? res.data ?? {}) }
    activeKey.value = keys.value[0] ?? ''
  } catch (e) {
    ElMessage.error('获取 ConfigMap 失败: ' + e.message)
  } finally {
    loading.value = false
  }
}

function addKey() {
  const k = newKey.value.trim()
  if (!k) return
  if (editData.value[k] !== undefined) {
    ElMessage.warning('键名已存在')
    return
  }
  editData.value[k] = ''
  activeKey.value   = k
  adding.value      = false
  newKey.value      = ''
}

async function removeKey(key) {
  await ElMessageBox.confirm(`确认删除键 "${key}"?`, '提示', { type: 'warning' })
  delete editData.value[key]
  activeKey.value = keys.value[0] ?? ''
}

async function save() {
  saving.value = true
  try {
    await updateConfigMap(props.namespace, props.cmName, editData.value)
    ElMessage.success('ConfigMap 已更新')
    visible.value = false
    emit('updated')
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.cm-layout {
  display: flex;
  gap: 0;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
  height: 420px;
}
.cm-keys {
  width: 200px;
  min-width: 200px;
  border-right: 1px solid #e4e7ed;
  background: #fafafa;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}
.cm-keys-title {
  font-size: 12px;
  color: #999;
  padding: 10px 12px 6px;
  font-weight: 600;
  letter-spacing: .5px;
}
.cm-key-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 13px;
  font-family: monospace;
  border-left: 3px solid transparent;
  transition: all .15s;
  word-break: break-all;
}
.cm-key-item:hover  { background: #f0f2f5; }
.cm-key-item.active { background: #ecf5ff; border-left-color: #409eff; color: #409eff; }
.cm-add-key { padding: 8px 12px; }
.cm-new-key { padding: 8px; display: flex; gap: 4px; }
.cm-value {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.cm-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ccc;
  font-size: 14px;
}
.cm-value-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
  background: #fff;
}
.cm-value-key {
  font-family: monospace;
  font-weight: 600;
  color: #409eff;
}
.cm-value :deep(.el-textarea__inner) {
  border: none;
  border-radius: 0;
  height: 100% !important;
  resize: none;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  line-height: 1.6;
}
</style>
