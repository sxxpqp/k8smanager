<template>
  <el-container style="height: 100vh;">
    <el-header style="background:#001529; display:flex; align-items:center; gap:12px;">
      <el-icon :size="24" color="#fff"><Platform /></el-icon>
      <span style="color:#fff; font-size:18px; font-weight:600;">K8s Pod 管理</span>
    </el-header>
    <el-main style="padding:0; background:#f0f2f5;">
      <PodManager />
    </el-main>
  </el-container>
</template>

<script setup>
import PodManager from './views/PodManager.vue'
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; }
</style>
