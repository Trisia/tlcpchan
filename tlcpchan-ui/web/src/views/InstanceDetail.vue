<template>
  <div class="instance-detail">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">{{ instance?.name }}</span>
        <el-tag :type="statusType(instance?.status || 'stopped')">{{ statusText(instance?.status || 'stopped') }}</el-tag>
      </template>
    </el-page-header>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
        <el-card>
          <template #header>
            <span>统计信息</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="6">
              <el-statistic title="总连接数" :value="stats?.connections_total || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="活跃连接" :value="stats?.connections_active || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="接收字节" :value="formatBytes(stats?.bytes_received || 0)" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="发送字节" :value="formatBytes(stats?.bytes_sent || 0)" />
            </el-col>
          </el-row>
        </el-card>

        <el-card style="margin-top: 20px">
          <template #header>
            <span>实例日志</span>
            <el-select v-model="logLevel" size="small" style="margin-left: 16px; width: 100px" @change="fetchLogs">
              <el-option label="全部" value="" />
              <el-option label="INFO" value="info" />
              <el-option label="WARN" value="warn" />
              <el-option label="ERROR" value="error" />
            </el-select>
          </template>
          <div class="log-container" v-loading="logsLoading">
            <div v-for="(log, i) in logs" :key="i" class="log-line">
              <span class="log-time">{{ log.time }}</span>
              <span :class="['log-level', log.level]">{{ log.level.toUpperCase() }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
            <el-empty v-if="logs.length === 0" description="暂无日志" />
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card>
          <template #header>
            <span>配置信息</span>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="类型">{{ instance?.type }}</el-descriptions-item>
            <el-descriptions-item label="协议">{{ instance?.protocol }}</el-descriptions-item>
            <el-descriptions-item label="认证模式">{{ authText(instance?.auth || 'none') }}</el-descriptions-item>
            <el-descriptions-item label="监听地址">{{ instance?.listen }}</el-descriptions-item>
            <el-descriptions-item label="目标地址">{{ instance?.target }}</el-descriptions-item>
            <el-descriptions-item label="运行时长">{{ formatUptime(instance?.uptime || 0) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>

         <el-card style="margin-top: 20px">
           <template #header>
             <span>健康检查</span>
           </template>
           <el-button type="success" @click="checkHealth" :loading="healthLoading">健康检查</el-button>
           <div v-if="healthResults" style="margin-top: 16px">
             <div v-for="result in healthResults.results" :key="result.protocol" style="margin-bottom: 12px">
               <div class="health-result-header">
                 <el-tag :type="result.success ? 'success' : 'danger'" size="small">
                   {{ result.protocol.toUpperCase() }}
                 </el-tag>
                  <span v-if="result.success" style="margin-left: 8px; color: #67c23a">
                    延迟: {{ result.latencyMs }}ms
                  </span>
                 <span v-else style="margin-left: 8px; color: #f56c6c">
                   失败: {{ result.error }}
                 </span>
               </div>
             </div>
           </div>
         </el-card>

         <el-card style="margin-top: 20px">
           <template #header>
             <span>操作</span>
           </template>
           <el-button type="primary" @click="start" :disabled="instance?.status === 'running'">启动</el-button>
           <el-button type="danger" @click="stop" :disabled="instance?.status !== 'running'">停止</el-button>
           <el-button type="warning" @click="reload" :disabled="instance?.status !== 'running'">重载</el-button>
         </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { instanceApi } from '@/api'
import type { Instance, InstanceHealthResponse } from '@/types'

const route = useRoute()
const router = useRouter()

const instance = ref<Instance | null>(null)
const stats = ref<{ connections_total: number; connections_active: number; bytes_received: number; bytes_sent: number } | null>(null)
const logs = ref<Array<{ time: string; level: string; message: string }>>([])
const logLevel = ref('')
const logsLoading = ref(false)
const healthLoading = ref(false)
const healthResults = ref<InstanceHealthResponse | null>(null)

const name = computed(() => route.params.name as string)

onMounted(async () => {
  await fetchInstance()
  await fetchStats()
  await fetchLogs()
})

async function fetchInstance() {
  instance.value = await instanceApi.get(name.value)
}

async function fetchStats() {
  stats.value = await instanceApi.stats(name.value)
}

async function fetchLogs() {
  logsLoading.value = true
  try {
    const data = await instanceApi.logs(name.value, 100, logLevel.value || undefined)
    logs.value = data.logs
  } finally {
    logsLoading.value = false
  }
}

async function checkHealth() {
  healthLoading.value = true
  healthResults.value = null
  try {
    healthResults.value = await instanceApi.health(name.value)
    ElMessage.success('健康检查完成')
  } catch (err) {
    ElMessage.error(`健康检查失败: ${(err as Error).message}`)
  } finally {
    healthLoading.value = false
  }
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / 1024 / 1024).toFixed(2) + ' MB'
  return (bytes / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  if (days > 0) return `${days}天 ${hours}小时`
  if (hours > 0) return `${hours}小时`
  return `${Math.floor(seconds / 60)}分钟`
}

function statusType(status: Instance['status']): '' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = { running: 'success', stopped: 'info', error: 'danger', created: 'warning' }
  return map[status] || ''
}

function statusText(status: Instance['status']): string {
  const map: Record<string, string> = { running: '运行中', stopped: '已停止', error: '错误', created: '已创建' }
  return map[status] || status
}

function authText(auth: Instance['auth']): string {
  const map: Record<string, string> = { none: '无', 'one-way': '单向', mutual: '双向' }
  return map[auth] || auth
}

async function start() {
  await instanceApi.start(name.value)
  await fetchInstance()
  ElMessage.success('实例已启动')
}

async function stop() {
  await instanceApi.stop(name.value)
  await fetchInstance()
  ElMessage.success('实例已停止')
}

async function reload() {
  await instanceApi.reload(name.value)
  await fetchInstance()
  ElMessage.success('实例已重载')
}
</script>

<style scoped>
.log-container {
  max-height: 400px;
  overflow-y: auto;
  background: #1d1e1f;
  padding: 12px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
}
.log-line {
  line-height: 1.8;
  color: #bfcbd9;
}
.log-time {
  color: #909399;
  margin-right: 8px;
}
.log-level {
  margin-right: 8px;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
}
.log-level.info {
  background: #409eff;
  color: #fff;
}
.log-level.warn {
  background: #e6a23c;
  color: #fff;
}
.log-level.error {
  background: #f56c6c;
  color: #fff;
}
.health-result-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>
