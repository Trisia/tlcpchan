<template>
  <div class="dashboard">
    <div class="welcome">
      <h2>TLCP Channel 管理面板</h2>
      <p>TLCP/TLS 协议代理工具，支持双协议并行工作</p>
    </div>

    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff">
              <el-icon size="28"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ health?.instances.total || 0 }}</div>
              <div class="stat-label">实例总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a">
              <el-icon size="28"><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ health?.instances.running || 0 }}</div>
              <div class="stat-label">运行中</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c">
              <el-icon size="28"><Key /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ health?.certificates.total || 0 }}</div>
              <div class="stat-label">证书总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-wrapper">
          <div class="stat-card">
            <div class="stat-icon" style="background: #909399">
              <el-icon size="28"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ formatUptime(info?.uptime || 0) }}</div>
              <div class="stat-label">运行时长</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="16">
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
          <el-table :data="instances" v-loading="instanceStore.loading" max-height="400">
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="protocol" label="协议" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.protocol === 'tlcp' ? 'primary' : 'success'">{{ row.protocol }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="listen" label="监听地址" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button v-if="row.status !== 'running'" type="primary" size="small" @click="start(row.name)">启动</el-button>
                <el-button v-else type="danger" size="small" @click="stop(row.name)">停止</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>系统信息</span>
          </template>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="版本">{{ info?.version }}</el-descriptions-item>
            <el-descriptions-item label="系统">{{ info?.os }}/{{ info?.arch }}</el-descriptions-item>
            <el-descriptions-item label="CPU核心数">{{ info?.numCpu }}</el-descriptions-item>
            <el-descriptions-item label="Goroutines">{{ info?.numGoroutine }}</el-descriptions-item>
            <el-descriptions-item label="内存">{{ info?.memAllocMb }} MB / {{ info?.memSysMb }} MB</el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card style="margin-top: 20px">
          <template #header>
            <span>快捷入口</span>
          </template>
          <div class="quick-links">
            <el-button type="primary" @click="$router.push('/instances')">
              <el-icon><Connection /></el-icon>
              实例管理
            </el-button>
            <el-button type="warning" @click="$router.push('/certificates')">
              <el-icon><Key /></el-icon>
              证书管理
            </el-button>
            <el-button type="info" @click="$router.push('/logs')">
              <el-icon><Document /></el-icon>
              日志查看
            </el-button>
            <el-button @click="$router.push('/settings')">
              <el-icon><Setting /></el-icon>
              系统设置
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useInstanceStore } from '@/stores/instance'
import { useSystemStore } from '@/stores/system'
import type { Instance } from '@/types'

const instanceStore = useInstanceStore()
const systemStore = useSystemStore()

const instances = computed(() => instanceStore.instances)
const info = computed(() => systemStore.info)
const health = computed(() => systemStore.health)

onMounted(async () => {
  await Promise.all([instanceStore.fetchInstances(), systemStore.fetchInfo(), systemStore.fetchHealth()])
})

function formatUptime(uptimeStr: string): string {
  return uptimeStr || '-'
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

async function start(name: string) {
  await instanceStore.startInstance(name)
  await systemStore.fetchHealth()
}

async function stop(name: string) {
  await instanceStore.stopInstance(name)
  await systemStore.fetchHealth()
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
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.quick-links {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}
.quick-links .el-button {
  width: 100%;
}
</style>
