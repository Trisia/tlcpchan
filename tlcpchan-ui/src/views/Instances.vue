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
        <el-form-item label="TLCP客户端认证" :disabled="form.protocol === 'tls'">
          <el-select v-model="form.tlcp!.clientAuthType" placeholder="请选择认证类型">
            <el-option label="不要求证书" value="no-client-cert" />
            <el-option label="请求证书" value="request-client-cert" />
            <el-option label="要求证书" value="require-any-client-cert" />
            <el-option label="验证已提供证书" value="verify-client-cert-if-given" />
            <el-option label="要求并验证证书" value="require-and-verify-client-cert" />
          </el-select>
        </el-form-item>
        <el-form-item label="TLS客户端认证" :disabled="form.protocol === 'tlcp'">
          <el-select v-model="form.tls!.clientAuthType" placeholder="请选择认证类型">
            <el-option label="不要求证书" value="no-client-cert" />
            <el-option label="请求证书" value="request-client-cert" />
            <el-option label="要求证书" value="require-any-client-cert" />
            <el-option label="验证已提供证书" value="verify-client-cert-if-given" />
            <el-option label="要求并验证证书" value="require-and-verify-client-cert" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择密钥">
          <el-select v-model="selectedKeystoreName" placeholder="请选择密钥（可选）" clearable @change="onKeystoreChange">
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
        
        <el-divider content-position="left">TLCP 高级配置</el-divider>
        <el-form-item label="最低版本">
          <el-select v-model="form.tlcp.minVersion" placeholder="请选择" :disabled="form.protocol === 'tls'">
            <el-option label="1.1" value="1.1" />
          </el-select>
        </el-form-item>
        <el-form-item label="最高版本">
          <el-select v-model="form.tlcp.maxVersion" placeholder="请选择" :disabled="form.protocol === 'tls'">
            <el-option label="1.1" value="1.1" />
          </el-select>
        </el-form-item>
        <el-form-item label="密码套件">
          <el-select v-model="form.tlcp.cipherSuites" placeholder="请选择" multiple :disabled="form.protocol === 'tls'">
            <el-option v-for="cs in TLCP_CIPHER_SUITES" :key="cs" :label="cs" :value="cs" />
          </el-select>
        </el-form-item>
        <el-form-item label="椭圆曲线">
          <el-select v-model="form.tlcp.curvePreferences" placeholder="请选择" multiple :disabled="form.protocol === 'tls' || selectedKeystoreType === 'RSA'">
            <el-option v-for="c in TLCP_CURVES" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="会话票据">
          <el-switch v-model="form.tlcp.sessionTickets" :disabled="form.protocol === 'tls'" />
        </El-form-item>
        <el-form-item label="会话缓存">
          <el-switch v-model="form.tlcp.sessionCache" :disabled="form.protocol === 'tls'" />
        </el-form-item>
        <el-form-item label="跳过证书验证">
          <el-switch v-model="form.tlcp.insecureSkipVerify" :disabled="form.protocol === 'tls'" />
        </el-form-item>

        <el-divider content-position="left">TLS 高级配置</el-divider>
        <el-form-item label="最低版本">
          <el-select v-model="form.tls.minVersion" placeholder="请选择" :disabled="form.protocol === 'tlcp'">
            <el-option label="1.0" value="1.0" />
            <el-option label="1.1" value="1.1" />
            <el-option label="1.2" value="1.2" />
            <el-option label="1.3" value="1.3" />
          </el-select>
        </el-form-item>
        <el-form-item label="最高版本">
          <el-select v-model="form.tls.maxVersion" placeholder="请选择" :disabled="form.protocol === 'tlcp'">
            <el-option label="1.0" value="1.0" />
            <el-option label="1.1" value="1.1" />
            <el-option label="1.2" value="1.2" />
            <el-option label="1.3" value="1.3" />
          </el-select>
        </el-form-item>
        <el-form-item label="密码套件">
          <el-select v-model="form.tls.cipherSuites" placeholder="请选择" multiple :disabled="form.protocol === 'tlcp'">
            <el-option v-for="cs in TLS_CIPHER_SUITES" :key="cs" :label="cs" :value="cs" />
          </el-select>
        </el-form-item>
        <el-form-item label="椭圆曲线">
          <el-select v-model="form.tls.curvePreferences" placeholder="请选择" multiple :disabled="form.protocol === 'tlcp'">
            <el-option v-for="c in TLS_CURVES" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
        <el-form-item label="会话票据">
          <el-switch v-model="form.tls.sessionTickets" :disabled="form.protocol === 'tlcp'" />
        </el-form-item>
        <el-form-item label="会话缓存">
          <el-switch v-model="form.tls.sessionCache" :disabled="form.protocol === 'tlcp'" />
        </el-form-item>
        <el-form-item label="跳过证书验证">
          <el-switch v-model="form.tls.insecureSkipVerify" :disabled="form.protocol === 'tlcp'" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="create" :loading="createLoading">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { instanceApi, keyStoreApi } from '@/api'
import type { Instance, InstanceConfig } from '@/types'

const router = useRouter()

const instances = ref<Instance[]>([])
const showCreateDialog = ref(false)
const keystores = ref<any[]>([])
const selectedKeystoreName = ref('')
const selectedKeystoreType = ref('')

// 按钮级别的加载状态
const instanceActions = ref<Record<string, boolean>>({})
const createLoading = ref(false)
const refreshLoading = ref(false)

const TLCP_CIPHER_SUITES = [
  'ECC_SM4_CBC_SM3',
  'ECC_SM4_GCM_SM3',
  'ECC_SM4_CCM_SM3',
  'ECDHE_SM4_CBC_SM3',
  'ECDHE_SM4_GCM_SM3',
  'ECDHE_SM4_CCM_SM3'
]

const TLS_CIPHER_SUITES = [
  'TLS_RSA_WITH_AES_128_GCM_SHA256',
  'TLS_RSA_WITH_AES_256_GCM_SHA384',
  'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256',
  'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384',
  'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256',
  'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384',
  'TLS_AES_128_GCM_SHA256',
  'TLS_AES_256_GCM_SHA384',
  'TLS_CHACHA20_POLY1305_SHA256'
]

const TLCP_CURVES = ['SM2']
const TLS_CURVES = ['P256', 'P38', 'P521', 'X25519']
const TLCP_VERSIONS = ['1.1']
const TLS_VERSIONS = ['1.0', '1.1', '1.2', '1.3']

const form = ref<Partial<InstanceConfig>>({
  name: '',
  type: 'server',
  protocol: 'auto',
  listen: ':443',
  target: '127.0.0.1:8080',
  enabled: true,
  tlcp: {
    clientAuthType: 'no-client-cert',
    minVersion: '1.1',
    maxVersion: '1.1',
    cipherSuites: [],
    curvePreferences: [],
    sessionTickets: false,
    sessionCache: false,
    insecureSkipVerify: false
  },
  tls: {
    clientAuthType: 'no-client-cert',
    minVersion: '1.2',
    maxVersion: '1.3',
    cipherSuites: [],
    curvePreferences: [],
    sessionTickets: false,
    sessionCache: false,
    insecureSkipVerify: false
  }
})

onMounted(() => {
  loadInstances()
  loadKeystores()
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

async function loadKeystores() {
  try {
    const result = await keyStoreApi.list()
    keystores.value = result.keystores || []
  } catch (err) {
    console.error('获取密钥列表失败:', err)
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

function onKeystoreChange(name: string) {
  if (name) {
    const ks = keystores.value.find(k => k.name === name)
    selectedKeystoreType.value = ks?.type || ''
  } else {
    selectedKeystoreType.value = ''
  }
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

async function create() {
  if (!form.value.name) {
    ElMessage.error('请输入实例名称')
    return
  }

  const data: any = { ...form.value }

  if (data.tlcp) {
    data.tlcp.auth = form.value.auth
  }
  if (data.tls) {
    data.tls.auth = form.value.auth
  }

  if (selectedKeystoreName.value) {
    const ksData = { name: selectedKeystoreName.value }
    if (form.value.protocol === 'tlcp' || form.value.protocol === 'auto') {
      data.tlcp = { ...data.tlcp, keystore: ksData }
    }
    if (form.value.protocol === 'tls' || form.value.protocol === 'auto') {
      data.tls = { ...data.tls, keystore: ksData }
    }
  }

  createLoading.value = true
  try {
    await instanceApi.create(data)
    showCreateDialog.value = false
    loadInstances()
    ElMessage.success('实例创建成功')
    form.value.name = ''
    selectedKeystoreName.value = ''
    form.value.tlcp = {
      minVersion: '1.1',
      maxVersion: '1.1',
      cipherSuites: [],
      curvePreferences: [],
      sessionTickets: false,
      sessionCache: false,
      insecureSkipVerify: false
    }
    form.value.tls = {
      minVersion: '1.2',
      maxVersion: '1.3',
      cipherSuites: [],
      curvePreferences: [],
      sessionTickets: false,
      sessionCache: false,
      insecureSkipVerify: false
    }
  } catch (err) {
    console.error('创建失败:', err)
    ElMessage.error('创建失败')
  } finally {
    createLoading.value = false
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
