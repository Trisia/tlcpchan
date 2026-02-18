<template>
  <div class="logs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>日志查看</span>
          <div>
            <el-select v-model="selectedInstance" placeholder="选择实例" style="width: 200px; margin-right: 12px" @change="fetchLogs">
              <el-option v-for="i in instances" :key="i.name" :label="i.name" :value="i.name" />
            </el-select>
            <el-select v-model="logLevel" placeholder="日志级别" style="width: 100px; margin-right: 12px" @change="fetchLogs">
              <el-option label="全部" value="" />
              <el-option label="DEBUG" value="debug" />
              <el-option label="INFO" value="info" />
              <el-option label="WARN" value="warn" />
              <el-option label="ERROR" value="error" />
            </el-select>
            <el-button type="primary" @click="fetchLogs">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <div class="log-container" v-loading="loading">
        <div v-for="(log, i) in logs" :key="i" class="log-line">
          <span class="log-time">{{ formatTime(log.time) }}</span>
          <span :class="['log-level', log.level]">{{ log.level.toUpperCase() }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
        <el-empty v-if="logs.length === 0" description="暂无日志" />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useInstanceStore } from '@/stores/instance'
import axios from 'axios'

const API_BASE = '/api/v1'
const store = useInstanceStore()

const instances = computed(() => store.instances)
const selectedInstance = ref('')
const logLevel = ref('')
const logs = ref<Array<{ time: string; level: string; message: string }>>([])
const loading = ref(false)

onMounted(() => {
  store.fetchInstances()
  setTimeout(() => {
    const first = instances.value[0]
    if (first) {
      selectedInstance.value = first.name
      fetchLogs()
    }
  }, 100)
})

function fetchLogs() {
  if (!selectedInstance.value) return
  loading.value = true
  const params = new URLSearchParams({ lines: String(500) })
  if (logLevel.value) params.append('level', logLevel.value)
  axios.get(`${API_BASE}/instances/${selectedInstance.value}/logs?${params.toString()}`)
    .then((res) => {
      logs.value = res.data.logs
    })
    .catch((err) => {
      console.error('获取日志失败', err)
    })
    .finally(() => {
      loading.value = false
    })
}

function formatTime(timeStr: string): string {
  return new Date(timeStr).toLocaleString('zh-CN')
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
}
.log-time {
  color: #909399;
  margin-right: 12px;
}
.log-level {
  margin-right: 12px;
  padding: 2px 8px;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
}
.log-level.debug {
  background: #606266;
  color: #fff;
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
</style>
