<template>
  <div class="instances">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>实例管理</span>
          <el-button type="primary" @click="router.push('/instances/create')">
            <el-icon><Plus /></el-icon>
            新建实例
          </el-button>
        </div>
      </template>

      <el-table :data="instances" v-loading="refreshLoading">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="config.type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ typeText(row.config.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="config.protocol" label="协议" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.config.protocol === 'tlcp' ? 'primary' : 'success'">{{ row.config.protocol }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="TLCP认证" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.config.tlcp?.clientAuthType || 'no-client-cert' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="TLS认证" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="success">{{ row.config.tls?.clientAuthType || 'no-client-cert' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="config.listen" label="监听地址" />
        <el-table-column prop="config.target" label="目标地址" />
        <el-table-column prop="enabled" label="启用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.enabled" @change="toggleEnabled(row)" />
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="viewDetail(row.name)">详情</el-button>
            <el-button v-if="row.status !== 'running'" type="success" size="small" link @click="start(row.name)" :loading="instanceActions[row.name]">启动</el-button>
            <el-button v-if="row.status === 'running'" type="warning" size="small" link @click="stop(row.name)" :loading="instanceActions[row.name]">停止</el-button>
            <el-button v-if="row.status === 'running'" type="info" size="small" link @click="reload(row.name)" :loading="instanceActions[row.name]">重载</el-button>
            <el-button type="danger" size="small" link @click="remove(row.name)" :loading="instanceActions[row.name]">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { instanceApi } from '@/api'
import type { Instance } from '@/types'

const router = useRouter()

const instances = ref<Instance[]>([])

const instanceActions = ref<Record<string, boolean>>({})
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

async function start(name: string) {
  instanceActions.value[name] = true
  try {
    await instanceApi.start(name)
    ElMessage.success('实例已启动')
    loadInstances()
  } catch (err) {
    console.error('启动失败:', err)
    ElMessage.error('启动失败')
  } finally {
    instanceActions.value[name] = false
  }
}

async function stop(name: string) {
  instanceActions.value[name] = true
  try {
    await instanceApi.stop(name)
    ElMessage.success('实例已停止')
    loadInstances()
  } catch (err) {
    console.error('停止失败:', err)
    ElMessage.error('停止失败')
  } finally {
    instanceActions.value[name] = false
  }
}

async function reload(name: string) {
  instanceActions.value[name] = true
  try {
    await instanceApi.reload(name)
    ElMessage.success('实例已重载')
    loadInstances()
  } catch (err) {
    console.error('重载失败:', err)
    ElMessage.error('重载失败')
  } finally {
    instanceActions.value[name] = false
  }
}

async function remove(name: string) {
  try {
    await ElMessageBox.confirm('确定要删除此实例吗？', '确认删除', { type: 'warning' })
    instanceActions.value[name] = true
    await instanceApi.delete(name)
    ElMessage.success('实例已删除')
    loadInstances()
  } catch (err) {
    if (err !== 'cancel') {
      console.error('删除失败:', err)
      ElMessage.error('删除失败')
    }
  } finally {
    instanceActions.value[name] = false
  }
}

async function toggleEnabled(row: Instance) {
  instanceActions.value[row.name] = true
  try {
    await instanceApi.update(row.name, { enabled: row.enabled })
    ElMessage.success(row.enabled ? '实例已启用' : '实例已禁用')
  } catch (err) {
    console.error('更新状态失败:', err)
    ElMessage.error('更新状态失败')
    row.enabled = !row.enabled
  } finally {
    instanceActions.value[row.name] = false
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
