<template>
  <div class="settings">
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>系统信息</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="版本">{{ info?.version }}</el-descriptions-item>
            <el-descriptions-item label="Go版本">{{ info?.go_version }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ info?.os }}/{{ info?.arch }}</el-descriptions-item>
            <el-descriptions-item label="启动时间">{{ formatTime(info?.start_time) }}</el-descriptions-item>
            <el-descriptions-item label="运行时长">{{ formatUptime(info?.uptime || 0) }}</el-descriptions-item>
            <el-descriptions-item label="进程ID">{{ info?.pid }}</el-descriptions-item>
            <el-descriptions-item label="Goroutines">{{ info?.goroutines }}</el-descriptions-item>
            <el-descriptions-item label="内存使用">{{ info?.memory.alloc_mb }} MB / {{ info?.memory.sys_mb }} MB</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span>健康状态</span>
          </template>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="状态">
              <el-tag :type="health?.status === 'healthy' ? 'success' : 'danger'">{{ health?.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="实例总数">{{ health?.instances.total }}</el-descriptions-item>
            <el-descriptions-item label="运行中">{{ health?.instances.running }}</el-descriptions-item>
            <el-descriptions-item label="已停止">{{ health?.instances.stopped }}</el-descriptions-item>
            <el-descriptions-item label="证书总数">{{ health?.certificates.total }}</el-descriptions-item>
            <el-descriptions-item label="已过期">
              <span :class="{ 'text-danger': health?.certificates.expired }">{{ health?.certificates.expired }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="即将过期">
              <span :class="{ 'text-warning': health?.certificates.expiring_soon }">{{ health?.certificates.expiring_soon }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card style="margin-top: 20px">
          <template #header>
            <span>操作</span>
          </template>
          <el-button type="primary" @click="reloadConfig">
            <el-icon><Refresh /></el-icon>
            重载配置
          </el-button>
          <el-button type="danger" @click="shutdown">
            <el-icon><SwitchButton /></el-icon>
            关闭服务
          </el-button>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSystemStore } from '@/stores/system'
import { configApi, systemApi } from '@/api'

const store = useSystemStore()

const info = computed(() => store.info)
const health = computed(() => store.health)

onMounted(() => {
  store.fetchInfo()
  store.fetchHealth()
})

function formatTime(timeStr?: string): string {
  if (!timeStr) return '-'
  return new Date(timeStr).toLocaleString('zh-CN')
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}天 ${hours}小时`
  if (hours > 0) return `${hours}小时 ${minutes}分钟`
  return `${minutes}分钟`
}

async function reloadConfig() {
  await configApi.reload()
  ElMessage.success('配置已重载')
}

async function shutdown() {
  await ElMessageBox.confirm('确定要关闭服务吗？此操作不可恢复。', '警告', { type: 'warning' })
  await systemApi.health()
  ElMessage.warning('服务正在关闭...')
}
</script>

<style scoped>
.text-danger {
  color: #f56c6c;
}
.text-warning {
  color: #e6a23c;
}
</style>
