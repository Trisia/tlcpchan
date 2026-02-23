<template>
  <div class="instance-detail">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">{{ instance?.name }}</span>
        <el-tag :type="statusType(instance?.status || 'stopped')">{{ statusText(instance?.status || 'stopped') }}</el-tag>
      </template>
    </el-page-header>

    <div style="margin-top: 20px">
      <!-- 控制区：操作 -->
      <el-card>
        <template #header>
          <span>操作</span>
        </template>
         <el-button type="primary" @click="start" :disabled="instance?.status === 'running'" :loading="actionLoading.start">启动</el-button>
         <el-button type="danger" @click="stop" :disabled="instance?.status !== 'running'" :loading="actionLoading.stop">停止</el-button>
         <el-button type="warning" @click="reload" :disabled="instance?.status !== 'running'" :loading="actionLoading.reload">重载</el-button>
         <el-button type="info" @click="restart" :loading="actionLoading.restart">重启</el-button>
         <el-button type="success" @click="edit" style="margin-left: 8px">编辑</el-button>
      </el-card>

      <!-- 控制区：健康检查 -->
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

      <!-- 基本信息 -->
      <el-card style="margin-top: 20px">
        <template #header>
          <span>基本信息</span>
        </template>
        <el-descriptions :column="4" border size="small">
          <el-descriptions-item label="类型">{{ instance?.config.type }}</el-descriptions-item>
          <el-descriptions-item label="协议">{{ instance?.config.protocol }}</el-descriptions-item>
          <el-descriptions-item label="监听地址">{{ instance?.config.listen }}</el-descriptions-item>
          <el-descriptions-item label="目标地址">{{ instance?.config.target }}</el-descriptions-item>
          <el-descriptions-item label="运行时长">{{ formatUptime(instance?.uptime || 0) }}</el-descriptions-item>
          <el-descriptions-item label="统计">
            <el-tag :type="instance?.config.stats?.enabled ? 'success' : 'info'" size="small">
              {{ instance?.config.stats?.enabled ? '已启用' : '已禁用' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 监控区：统计信息 -->
      <el-card style="margin-top: 20px">
        <template #header>
          <span>统计信息</span>
        </template>
         <el-row :gutter="20">
           <el-col :span="6">
             <el-statistic title="总连接数" :value="stats?.totalConnections || 0" />
           </el-col>
           <el-col :span="6">
             <el-statistic title="活跃连接" :value="stats?.activeConnections || 0" />
           </el-col>
            <el-col :span="6">
              <el-statistic title="接收字节" :value="stats?.bytesReceived || 0" :formatter="formatBytes" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="发送字节" :value="stats?.bytesSent || 0" :formatter="formatBytes" />
            </el-col>
         </el-row>
      </el-card>

      <!-- 协议配置区 -->
      <el-card style="margin-top: 20px">
        <template #header>
          <span>协议配置</span>
        </template>
        <el-collapse v-model="activeCollapse" accordion>
          <el-collapse-item name="tlcp" title="TLCP 配置" :disabled="instance?.config.protocol === 'tls'">
            <ProtocolConfigDetail
              v-if="instance?.config.tlcp"
              :config="instance.config.tlcp"
              :is-tlcp="true"
            />
          </el-collapse-item>
          <el-collapse-item name="tls" title="TLS 配置" :disabled="instance?.config.protocol === 'tlcp'">
            <ProtocolConfigDetail
              v-if="instance?.config.tls"
              :config="instance.config.tls"
              :is-tlcp="false"
            />
          </el-collapse-item>
        </el-collapse>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ProtocolConfigDetail from '@/components/ProtocolConfigDetail.vue'
import { instanceApi } from '@/api'
import type { Instance, InstanceHealthResponse, InstanceStats } from '@/types'

const route = useRoute()
const router = useRouter()

const instance = ref<Instance | null>(null)
const stats = ref<InstanceStats | null>(null)
const logs = ref<Array<{ time: string; level: string; message: string }>>([])
const logLevel = ref('')
const logsLoading = ref(false)
const healthLoading = ref(false)
const healthResults = ref<InstanceHealthResponse | null>(null)
const activeCollapse = ref(['tlcp', 'tls'])

const name = computed(() => route.params.name as string)

onMounted(() => {
  fetchInstance()
  fetchStats()
  fetchLogs()
})

async function fetchInstance() {
  try {
    instance.value = await instanceApi.get(name.value)
  } catch (err) {
    console.error('获取实例失败:', err)
  }
}

async function fetchStats() {
  try {
    stats.value = await instanceApi.stats(name.value)
  } catch (err) {
    console.error('获取统计失败:', err)
  }
}

async function fetchLogs() {
  logsLoading.value = true
  try {
    const params: any = { lines: 100 }
    if (logLevel.value) params.level = logLevel.value
    logs.value = await instanceApi.logs(name.value, params)
  } catch (err) {
    console.error('获取日志失败:', err)
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
  } catch (err: any) {
    ElMessage.error(`健康检查失败: ${err.message}`)
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

const actionLoading = ref<Record<string, boolean>>({})

function edit() {
  router.push(`/instances/${name.value}/edit`)
}

async function start() {
  actionLoading.value.start = true
  try {
    await instanceApi.start(name.value)
    fetchInstance()
    ElMessage.success('实例已启动')
  } catch (err) {
    console.error('启动失败:', err)
    ElMessage.error('启动失败')
  } finally {
    actionLoading.value.start = false
  }
}

async function stop() {
  actionLoading.value.stop = true
  try {
    await instanceApi.stop(name.value)
    fetchInstance()
    ElMessage.success('实例已停止')
  } catch (err) {
    console.error('停止失败:', err)
    ElMessage.error('停止失败')
  } finally {
    actionLoading.value.stop = false
  }
}

async function reload() {
  actionLoading.value.reload = true
  try {
    await instanceApi.reload(name.value)
    fetchInstance()
    ElMessage.success('实例已重载')
  } catch (err) {
    console.error('重载失败:', err)
    ElMessage.error('重载失败')
  } finally {
    actionLoading.value.reload = false
  }
}

async function restart() {
  actionLoading.value.restart = true
  try {
    await instanceApi.restart(name.value)
    fetchInstance()
    ElMessage.success('实例已重启')
  } catch (err) {
    console.error('重启失败:', err)
    ElMessage.error('重启失败')
  } finally {
    actionLoading.value.restart = false
  }
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
