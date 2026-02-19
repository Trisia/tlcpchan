<template>
  <div class="instances">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>实例管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            新建实例
          </el-button>
        </div>
      </template>

      <el-table :data="instances" v-loading="loading">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ typeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="protocol" label="协议" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.protocol === 'tlcp' ? 'primary' : 'success'">{{ row.protocol }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="auth" label="认证模式" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ authText(row.auth) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="listen" label="监听地址" />
        <el-table-column prop="target" label="目标地址" />
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
            <el-button v-if="row.status !== 'running'" type="success" size="small" link @click="start(row.name)">启动</el-button>
            <el-button v-if="row.status === 'running'" type="warning" size="small" link @click="stop(row.name)">停止</el-button>
            <el-button v-if="row.status === 'running'" type="info" size="small" link @click="reload(row.name)">重载</el-button>
            <el-button type="danger" size="small" link @click="remove(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showCreateDialog" title="新建实例" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="实例名称" required>
          <el-input v-model="form.name" placeholder="请输入实例名称" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="form.type" placeholder="请选择类型">
            <el-option label="服务端代理" value="server" />
            <el-option label="客户端代理" value="client" />
            <el-option label="HTTP服务端" value="http-server" />
            <el-option label="HTTP客户端" value="http-client" />
          </el-select>
        </el-form-item>
        <el-form-item label="协议" required>
          <el-select v-model="form.protocol" placeholder="请选择协议">
            <el-option label="自动" value="auto" />
            <el-option label="TLCP" value="tlcp" />
            <el-option label="TLS" value="tls" />
          </el-select>
        </el-form-item>
        <el-form-item label="认证模式">
          <el-select v-model="form.auth" placeholder="请选择认证模式">
            <el-option label="无认证" value="none" />
            <el-option label="单向认证" value="one-way" />
            <el-option label="双向认证" value="mutual" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择密钥">
          <el-select v-model="selectedKeystoreName" placeholder="请选择密钥（可选）" clearable>
            <el-option
              v-for="ks in keystores"
              :key="ks.name"
              :label="ks.name"
              :value="ks.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="监听地址" required>
          <el-input v-model="form.listen" placeholder=":443" />
        </el-form-item>
        <el-form-item label="目标地址" required>
          <el-input v-model="form.target" placeholder="127.0.0.1:8080" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="create">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useInstanceStore } from '@/stores/instance'
import { instanceApi, keyStoreApi } from '@/api'
import type { Instance } from '@/types'

const router = useRouter()
const store = useInstanceStore()

const loading = computed(() => store.loading)
const instances = computed(() => store.instances)
const showCreateDialog = ref(false)
const keystores = ref<any[]>([])
const selectedKeystoreName = ref('')

const form = ref<Partial<Instance>>({
  name: '',
  type: 'server',
  protocol: 'auto',
  auth: 'none',
  listen: ':443',
  target: '127.0.0.1:8080',
  enabled: true,
})

onMounted(async () => {
  await store.fetchInstances()
  await fetchKeystores()
})

async function fetchKeystores() {
  try {
    keystores.value = await keyStoreApi.list() || []
  } catch (err) {
    console.error('获取密钥列表失败', err)
  }
}

function typeText(type: Instance['type']): string {
  const map: Record<string, string> = { server: '服务端', client: '客户端', 'http-server': 'HTTP服务端', 'http-client': 'HTTP客户端' }
  return map[type] || type
}

function authText(auth: Instance['auth']): string {
  const map: Record<string, string> = { none: '无', 'one-way': '单向', mutual: '双向' }
  return map[auth] || auth
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
  await store.startInstance(name)
  ElMessage.success('实例已启动')
}

async function stop(name: string) {
  await store.stopInstance(name)
  ElMessage.success('实例已停止')
}

async function reload(name: string) {
  await instanceApi.reload(name)
  await store.fetchInstances()
  ElMessage.success('实例已重载')
}

async function remove(name: string) {
  await ElMessageBox.confirm('确定要删除此实例吗？', '确认删除', { type: 'warning' })
  await store.deleteInstance(name)
  ElMessage.success('实例已删除')
}

async function toggleEnabled(row: Instance) {
  await instanceApi.update(row.name, { enabled: row.enabled })
  ElMessage.success(row.enabled ? '实例已启用' : '实例已禁用')
}

async function create() {
  if (!form.value.name) {
    ElMessage.error('请输入实例名称')
    return
  }

  const data: any = { ...form.value }

  if (selectedKeystoreName.value) {
    const ksData = { name: selectedKeystoreName.value }
    if (form.value.protocol === 'tlcp' || form.value.protocol === 'auto') {
      data.tlcp = { ...data.tlcp, keystore: ksData }
    }
    if (form.value.protocol === 'tls' || form.value.protocol === 'auto') {
      data.tls = { ...data.tls, keystore: ksData }
    }
  }

  await instanceApi.create(data)
  showCreateDialog.value = false
  await store.fetchInstances()
  ElMessage.success('实例创建成功')
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
