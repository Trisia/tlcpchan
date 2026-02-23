<template>
  <div class="create-instance">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">创建实例</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <el-form :model="form" label-width="120px" v-loading="loading">
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
          <el-select v-model="form.tlcp.clientAuthType" placeholder="请选择认证类型">
            <el-option label="不要求证书" value="no-client-cert" />
            <el-option label="请求证书" value="request-client-cert" />
            <el-option label="要求证书" value="require-any-client-cert" />
            <el-option label="验证已提供证书" value="verify-client-cert-if-given" />
            <el-option label="要求并验证证书" value="require-and-verify-client-cert" />
          </el-select>
        </el-form-item>
        <el-form-item label="TLS客户端认证" :disabled="form.protocol === 'tlcp'">
          <el-select v-model="form.tls.clientAuthType" placeholder="请选择认证类型">
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
        </el-form-item>
        <el-form-item label="会话缓存">
          <el-switch v-model="form.tlcp.sessionCache" :disabled="form.protocol === 'tls'" />
        </el-form-item>
        <el-form-item label="跳过会话证书验证">
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

        <el-form-item>
          <el-button type="primary" @click="create" :loading="loading">创建</el-button>
          <el-button @click="router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { instanceApi, keyStoreApi } from '@/api'
import type { InstanceConfig } from '@/types'

const router = useRouter()

const loading = ref(false)
const keystores = ref<any[]>([])
const selectedKeystoreName = ref('')
const selectedKeystoreType = ref('')

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

const form = ref<InstanceConfig>({
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
    insecureSkipVerify: false,
    keystore: undefined
  },
  tls: {
    clientAuthType: 'no-client-cert',
    minVersion: '1.2',
    maxVersion: '1.3',
    cipherSuites: [],
    curvePreferences: [],
    sessionTickets: false,
    sessionCache: false,
    insecureSkipVerify: false,
    keystore: undefined
  }
})

onMounted(() => {
  loadKeystores()
})

async function loadKeystores() {
  try {
    const result = await keyStoreApi.list()
    keystores.value = result.keystores || []
  } catch (err) {
    console.error('获取密钥列表失败:', err)
  }
}

function onKeystoreChange(name: string) {
  if (name) {
    const ks = keystores.value.find(k => k.name === name)
    selectedKeystoreType.value = ks?.type || ''
  } else {
    selectedKeystoreType.value = ''
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

  loading.value = true
  try {
    await instanceApi.create(data)
    ElMessage.success('实例创建成功')
    router.push('/instances')
  } catch (err: any) {
    console.error('创建失败:', err)
    ElMessage.error('创建失败: ' + (err.response?.data || err.message))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.text-large {
  font-size: 18px;
}
</style>
