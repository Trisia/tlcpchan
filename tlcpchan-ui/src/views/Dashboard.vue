<template>
  <div class="dashboard">
    <div class="welcome">
      <h2>TLCP Channel 管理面板</h2>
      <p>TLCP/TLS 协议代理工具，支持双协议并行工作</p>
    </div>

    <el-row :gutter="20">
      <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff">
              <el-icon size="28"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ instances.length }}</div>
              <div class="stat-label">实例总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a">
              <el-icon size="28"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ runningCount }}</div>
              <div class="stat-label">运行中</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c">
              <el-icon size="28"><DataLine /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatBytes(totalBytesReceived) }}</div>
              <div class="stat-label">总接收</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #909399">
              <el-icon size="28"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatUptime(info?.uptime) }}</div>
              <div class="stat-label">运行时长</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="content-row">
      <el-col :xs="24" :sm="24" :md="24" :lg="24" :xl="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>实例状态</span>
              <el-button type="primary" size="small" @click="$router.push('/instances')">
                管理实例
                <el-icon class="el-icon--right"><ArrowRight /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="table-container">
            <el-table :data="instances" v-loading="instanceLoading" max-height="400">
              <el-table-column prop="name" label="名称" />
               <el-table-column prop="config.type" label="类型" width="100" class="hide-on-mobile">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.config.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="config.protocol" label="协议" width="80">
                <template #default="{ row }">
                  <el-tag size="small" :type="row.config.protocol === 'tlcp' ? 'primary' : 'success'">{{ row.config.protocol }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="config.listen" label="监听地址" class="hide-on-mobile" />
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { systemApi, instanceApi } from '@/api'
import type { Instance, SystemInfo, InstanceStats } from '@/types'
import { Connection, CircleCheck, DataLine, Timer, ArrowRight } from '@element-plus/icons-vue'

const instances = ref<Instance[]>([])
const instanceStats = ref<Record<string, InstanceStats>>({})
const info = ref<SystemInfo | null>(null)
const instanceLoading = ref(false)

const runningCount = computed(() => instances.value.filter(i => i.status === 'running').length)
const totalBytesReceived = computed(() => Object.values(instanceStats.value).reduce((sum, s) => sum + s.bytesReceived, 0))

onMounted(async () => {
  await Promise.all([fetchInstances(), fetchInfo()])
})

async function fetchInstances() {
  instanceLoading.value = true
  try {
    instances.value = await instanceApi.list()
    await fetchInstanceStats()
  } finally {
    instanceLoading.value = false
  }
}

async function fetchInstanceStats() {
  const statsPromises = instances.value.map(async (inst) => {
    try {
      const stats = await instanceApi.stats(inst.name)
      return { name: inst.name, stats }
    } catch (error) {
      console.error(`获取实例 ${inst.name} 统计失败:`, error)
      return null
    }
  })

  const results = await Promise.all(statsPromises)
  instanceStats.value = {}
  for (const result of results) {
    if (result) {
      instanceStats.value[result.name] = result.stats
    }
  }
}

async function fetchInfo() {
  try {
    info.value = await systemApi.info()
  } catch (error) {
    console.error('获取系统信息失败:', error)
  }
}

function formatUptime(uptimeStr?: string): string {
  if (!uptimeStr) return '-'
  return uptimeStr
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}

function statusType(status: Instance['status']): '' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    running: 'success',
    stopped: 'info',
    error: 'danger',
    created: 'warning',
  }
  return map[status] || ''
}

function statusText(status: Instance['status']): string {
  const map: Record<string, string> = { running: '运行中', stopped: '已停止', error: '错误', created: '已创建' }
  return map[status] || status
}

</script>

<style scoped>
.dashboard {
  width: 100%;
}
.welcome {
  margin-bottom: 20px;
}
.welcome h2 {
  margin: 0 0 8px 0;
  font-size: 22px;
  color: #303133;
}
.welcome p {
  margin: 0;
  color: #909399;
  font-size: 14px;
}
.stat-card-wrapper {
  cursor: pointer;
  transition: transform 0.2s;
  margin-bottom: 20px;
}
.stat-card-wrapper:hover {
  transform: translateY(-2px);
}
.stat-card {
  display: flex;
  align-items: center;
}
.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}
.stat-info {
  margin-left: 16px;
}
.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}
.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}
.content-row {
  margin-top: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.table-container {
  overflow-x: auto;
}
.quick-links-card {
  margin-top: 20px;
}
.quick-links {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.quick-links .el-button {
  width: 100%;
}

@media (max-width: 768px) {
  .welcome h2 {
    font-size: 18px;
  }
  .stat-value {
    font-size: 20px;
  }
  .stat-icon {
    width: 48px;
    height: 48px;
  }
  .hide-on-mobile {
    display: none;
  }
  .quick-links {
    grid-template-columns: 1fr;
  }
}
</style>
