<template>
  <el-dialog
    v-model="visible"
    :title="`Secret — ${secretName}`"
    width="600px"
    @open="fetchDetail"
  >
    <div v-loading="loading" style="min-height:120px">
      <el-descriptions :column="1" border size="small" v-if="detail">
        <el-descriptions-item label="类型">
          <el-tag size="small" type="info">{{ detail.type }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider>键值列表</el-divider>

      <el-table :data="keyRows" size="small" style="width:100%">
        <el-table-column label="Key" prop="key" width="200">
          <template #default="{ row }">
            <span style="font-family:monospace; color:#409eff">{{ row.key }}</span>
          </template>
        </el-table-column>

        <el-table-column label="Value">
          <template #default="{ row }">
            <div class="secret-val">
              <span v-if="!row.visible" class="masked">{{ '●'.repeat(12) }}</span>
              <span v-else class="secret-text">{{ row.value }}</span>
              <el-button link size="small" :icon="row.visible ? Hide : View"
                @click="row.visible = !row.visible" />
              <el-button link size="small" :icon="CopyDocument"
                @click="copy(row.value)" />
            </div>
          </template>
        </el-table-column>

        <el-table-column label="字节数" width="80" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small">{{ row.size }}B</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <template #footer>
      <el-alert type="warning" :closable="false" show-icon
        title="Secret 内容敏感，请勿在不安全环境下展示" style="margin-bottom:8px" />
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { View, Hide, CopyDocument } from '@element-plus/icons-vue'
import { getSecretDetail } from '../api/index.js'

const props = defineProps({
  modelValue: Boolean,
  namespace:  String,
  secretName: String
})
const emit = defineEmits(['update:modelValue'])

const visible = ref(false)
const loading = ref(false)
const detail  = ref(null)

watch(() => props.modelValue, v => { visible.value = v })
watch(visible, v => emit('update:modelValue', v))

// 每行展示一个 key，带 visible 状态
const keyRows = computed(() => {
  if (!detail.value?.data) return []
  return Object.entries(detail.value.data).map(([key, val]) => ({
    key,
    value: val,
    size:  val?.length ?? 0,
    visible: false
  }))
})

async function fetchDetail() {
  loading.value = true
  detail.value  = null
  try {
    const res = await getSecretDetail(props.namespace, props.secretName)
    detail.value = res.data ?? res
  } catch (e) {
    ElMessage.error('获取 Secret 失败: ' + e.message)
  } finally {
    loading.value = false
  }
}

function copy(val) {
  navigator.clipboard.writeText(val)
    .then(() => ElMessage.success('已复制'))
    .catch(() => ElMessage.error('复制失败'))
}
</script>

<style scoped>
.secret-val {
  display: flex;
  align-items: center;
  gap: 4px;
}
.masked {
  letter-spacing: 2px;
  color: #999;
  font-size: 14px;
}
.secret-text {
  font-family: monospace;
  font-size: 12px;
  word-break: break-all;
  color: #303133;
}
</style>
