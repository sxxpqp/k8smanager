<template>
  <el-dialog
    v-model="visible"
    title="更新镜像"
    width="520px"
    @open="onOpen"
  >
    <el-form :model="form" label-width="80px" @submit.prevent>
      <!-- 容器选择 -->
      <el-form-item label="容器">
        <el-select v-model="form.container" style="width:100%" @change="onContainerChange">
          <el-option
            v-for="c in deployment?.containers ?? []"
            :key="c.name"
            :label="c.name"
            :value="c.name"
          >
            <div style="display:flex; justify-content:space-between; gap:16px">
              <span>{{ c.name }}</span>
              <span style="color:#999; font-size:12px; font-family:monospace">{{ c.image }}</span>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <!-- 当前镜像（只读） -->
      <el-form-item label="当前镜像">
        <el-input :value="currentImage" readonly>
          <template #prefix><el-icon><Picture /></el-icon></template>
        </el-input>
      </el-form-item>

      <!-- 新镜像 -->
      <el-form-item label="新镜像">
        <el-input
          v-model="form.image"
          placeholder="如：nginx:1.26 或 myrepo/app:v2.0.1"
          clearable
          autofocus
        >
          <template #prefix><el-icon><Upload /></el-icon></template>
        </el-input>
      </el-form-item>

      <!-- 快捷 Tag -->
      <el-form-item label="">
        <div class="tag-tips">
          <span class="tip-label">常用 tag：</span>
          <el-tag
            v-for="tag in quickTags"
            :key="tag"
            class="tag-item"
            size="small"
            effect="plain"
            style="cursor:pointer"
            @click="applyTag(tag)"
          >{{ tag }}</el-tag>
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        type="primary"
        :loading="loading"
        :disabled="!form.image || form.image === currentImage"
        @click="submit"
      >确认更新</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture, Upload } from '@element-plus/icons-vue'
import { updateImage } from '../api/index.js'

const props = defineProps({
  modelValue: Boolean,
  // deployment 对象：{ name, namespace, containers: [{name, image}] }
  deployment:  Object
})
const emit = defineEmits(['update:modelValue', 'updated'])

const visible = ref(false)
const loading = ref(false)
const form    = ref({ container: '', image: '' })

watch(() => props.modelValue, v => { visible.value = v })
watch(visible, v => emit('update:modelValue', v))

// 当前选中容器的镜像
const currentImage = computed(() => {
  const c = props.deployment?.containers?.find(c => c.name === form.value.container)
  return c?.image ?? ''
})

// 根据当前镜像提取 repo，生成快捷 tag
const quickTags = computed(() => {
  const img = currentImage.value
  if (!img) return []
  const repo = img.includes(':') ? img.split(':')[0] : img
  return [`${repo}:latest`, `${repo}:stable`, `${repo}:v2`]
})

function onOpen() {
  const first = props.deployment?.containers?.[0]
  form.value.container = first?.name ?? ''
  form.value.image     = first?.image ?? ''
}

function onContainerChange() {
  form.value.image = currentImage.value
}

function applyTag(tag) {
  form.value.image = tag
}

async function submit() {
  if (!form.value.image) return
  loading.value = true
  try {
    await updateImage(
      props.deployment.namespace,
      props.deployment.name,
      form.value.container,
      form.value.image
    )
    ElMessage.success(`镜像已更新为 ${form.value.image}，滚动更新中...`)
    visible.value = false
    emit('updated')
  } catch (e) {
    ElMessage.error(e.message)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.tag-tips { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.tip-label { font-size: 12px; color: #999; white-space: nowrap; }
.tag-item  { transition: all .2s; }
.tag-item:hover { color: #409eff; border-color: #409eff; }
</style>
