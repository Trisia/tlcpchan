<template>
  <div class="logs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>工作日志</span>
          <div class="header-actions">
            <el-select v-model="logLevel" placeholder="日志级别" style="width: 100px; margin-right: 12px" @change="fetchLogs">
              <el-option label="全部" value="" />
              <el-option label="DEBUG" value="debug" />
              <el-option label="INFO" value="info" />
              <el-option label="WARN" value="warn" />
              <el-option label="ERROR" value="error" />
            </el-select>
            <el-button type="primary" @click="fetchLogs" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button type="success" @click="downloadCurrent">
              <el-icon><Download /></el-icon>
              下载当前
            </el-button>
            <el-button type="warning" @click="downloadAll">
              <el-icon><FolderOpened /></el-icon>
              下载全部
            </el-button>
            <el-button @click="toggleAutoRefresh" :type="autoRefresh ? 'success' : 'default'">
              <el-icon><Timer /></el-icon>
              {{ autoRefresh ? '停止刷新' : '自动刷新' }}
            </el-button>
          </div>
        </div>
      </template>

      <div class="log-stats" v-if="logStats">
        <span>文件: {{ logStats.file }}</span>
        <span>总行数: {{ logStats.total }}</span>
        <span>显示: {{ logStats.returned }}行</span>
      </div>

      <div class="log-container" v-loading="loading">
        <div v-for="(log, i) in logs" :key="i" class="log-line">
          <span class="log-message">{{ log }}</span>
        </div>
        <el-empty v-if="logs.length === 0" description="暂无日志" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { logsApi } from '@/api'
import { ElMessage } from 'element-plus'

const logs = ref<string[]>([])
const logLevel = ref('')
const loading = ref(false)
const autoRefresh = ref(false)
const logStats = ref<{ file: string; total: number; returned: number } | null>(null)
let refreshTimer: ReturnType<typeof setInterval> | null = null

onMounted(async () => {
  await fetchLogs()
})

onUnmounted(() => {
  stopAutoRefresh()
})

async function fetchLogs() {
  loading.value = true
  try {
    const result = await logsApi.content({ lines: 500, level: logLevel.value })
    logs.value = result.lines || []
    logStats.value = {
      file: result.file || '',
      total: result.total || 0,
      returned: result.returned || 0
    }
  } catch (err: any) {
    ElMessage.error('获取日志失败: ' + (err.response?.data || err.message))
  } finally {
    loading.value = false
  }
}

async function downloadCurrent() {
  try {
    const fileName = logStats.value?.file || 'tlcpchan.log'
    await logsApi.download(fileName)
    ElMessage.success('下载成功')
  } catch (err: any) {
    ElMessage.error('下载失败: ' + (err.response?.data || err.message))
  }
}

async function downloadAll() {
  try {
    await logsApi.downloadAll()
    ElMessage.success('下载成功')
  } catch (err: any) {
    ElMessage.error('下载失败: ' + (err.response?.data || err.message))
  }
}

function toggleAutoRefresh() {
  if (autoRefresh.value) {
    stopAutoRefresh()
    ElMessage.info('已停止自动刷新')
  } else {
    startAutoRefresh()
    ElMessage.success('已开启自动刷新（每5秒）')
  }
}

function startAutoRefresh() {
  stopAutoRefresh()
  autoRefresh.value = true
  refreshTimer = setInterval(() => {
    fetchLogs()
  }, 5000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
  autoRefresh.value = false
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.log-stats {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 4px;
  margin-bottom: 16px;
  font-size: 14px;
  color: #606266;
}

.log-stats span {
  margin-right: 24px;
}

.log-container {
  max-height: 600px;
  overflow-y: auto;
  background: #1d1e1f;
  padding: 16px;
  border-radius: 4px;
  font-family: 'Fira Code', 'Monaco', 'Consolas', monospace;
  font-size: 13px;
}

.log-line {
  line-height: 1.8;
  color: #bfcbd9;
  word-break: break-all;
}

.log-message {
  white-space: pre-wrap;
}
</style>
