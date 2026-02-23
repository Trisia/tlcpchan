<template>
  <div class="instances">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>实例管理</span>
          <el-button type="primary" @click="router.push('/instances/create')" class="create-btn">
            <el-icon><Plus /></el-icon>
            新建实例
          </el-button>
        </div>
      </template>

      <div class="table-container">
        <el-table :data="instances" v-loading="refreshLoading">
          <el-table-column prop="name" label="名称" min-width="120" />
          <el-table-column prop="config.type" label="类型" width="120" class-name="hide-on-mobile">
            <template #default="{ row }">
              <el-tag size="small">{{ typeText(row.config.type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="config.protocol" label="协议" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.config.protocol === 'tlcp' ? 'primary' : 'success'">{{ row.config.protocol }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="TLCP认证" width="150" class-name="hide-on-tablet">
            <template #default="{ row }">
              <el-tag size="small" type="info">{{ row.config.tlcp?.clientAuthType || 'no-client-cert' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="TLS认证" width="150" class-name="hide-on-tablet">
            <template #default="{ row }">
              <el-tag size="small" type="success">{{ row.config.tls?.clientAuthType || 'no-client-cert' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="config.listen" label="监听地址" class-name="hide-on-mobile" />
          <el-table-column prop="config.target" label="目标地址" class-name="hide-on-mobile" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" size="small" link @click="viewDetail(row.name)">进入管理</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { instanceApi } from '@/api'
import type { Instance } from '@/types'

const router = useRouter()

const instances = ref<Instance[]>([])

const refreshLoading = ref(false)

onMounted(() => {
  loadInstances()
})

async function loadInstances() {
  refreshLoading.value = true
  try {
    instances.value = await instanceApi.list()
  } catch (err) {
    console.error('加载实例失败:', err)
  } finally {
    refreshLoading.value = false
  }
}

function typeText(type: Instance['config']['type']): string {
  const map: Record<string, string> = { server: '服务端', client: '客户端', 'http-server': 'HTTP服务端', 'http-client': 'HTTP客户端' }
  return map[type] || type
}

function statusType(status: Instance['status']): '' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = { running: 'success', stopped: 'info', error: 'danger', created: 'warning' }
  return map[status] || ''
}

function statusText(status: Instance['status']): string {
  const map: Record<string, string> = { running: '运行中', stopped: '已停止', error: '错误', created: '已创建' }
  return map[status] || status
}

function viewDetail(name: string) {
  router.push(`/instances/${name}`)
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
