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
            <el-option label="HTTP服务" value="http-server" />
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
        <el-form-item label="监听地址" required>
          <el-input v-model="form.listen" placeholder=":443" />
        </el-form-item>
        <el-form-item label="目标地址" required>
          <el-input v-model="form.target" placeholder="127.0.0.1:8080" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="create" :loading="loading">创建</el-button>
          <el-button @click="router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- TLCP 配置卡片 -->
    <el-card style="margin-top: 20px" v-if="form.protocol !== 'tls'">
      <template #header>
        <span>TLCP 配置 {{ form.protocol === 'tlcp' ? '(必填)' : '' }}</span>
      </template>
      <el-form :model="form" label-width="140px">
        <KeystoreConfig
          v-model="form.tlcp.keystoreConfig"
          :keystores="keystores"
          :is-tlcp="true"
          :required="form.protocol === 'tlcp'"
        />
        <el-form-item label="客户端认证类型">
          <el-select v-model="form.tlcp.clientAuthType" placeholder="请选择认证类型">
            <el-option label="no-client-cert" value="no-client-cert" />
            <el-option label="request-client-cert" value="request-client-cert" />
            <el-option label="require-any-client-cert" value="require-any-client-cert" />
            <el-option label="verify-client-cert-if-given" value="verify-client-cert-if-given" />
            <el-option label="require-and-verify-client-cert" value="require-and-verify-client-cert" />
          </el-select>
        </el-form-item>
        <el-form-item label="最低版本">
          <el-select v-model="form.tlcp.minVersion" placeholder="请选择">
            <el-option label="1.1" value="1.1" />
          </el-select>
        </el-form-item>
        <el-form-item label="最高版本">
          <el-select v-model="form.tlcp.maxVersion" placeholder="请选择">
            <el-option label="1.1" value="1.1" />
          </el-select>
        </el-form-item>
        <el-form-item label="密码套件">
          <el-checkbox-group v-model="form.tlcp.cipherSuites">
            <div class="cipher-grid">
              <el-checkbox v-for="cs in TLCP_CIPHER_SUITES" :key="cs" :label="cs" :value="cs" />
            </div>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="握手重用">
          <el-switch v-model="form.tlcp.sessionCache" />
        </el-form-item>
        <el-form-item label="跳过证书验证">
          <el-switch v-model="form.tlcp.insecureSkipVerify" />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- TLS 配置卡片 -->
    <el-card style="margin-top: 20px" v-if="form.protocol !== 'tlcp'">
      <template #header>
        <span>TLS 配置 {{ form.protocol === 'tls' ? '(必填)' : '' }}</span>
      </template>
      <el-form :model="form" label-width="140px">
        <KeystoreConfig
          v-model="form.tls.keystoreConfig"
          :keystores="keystores"
          :is-tlcp="false"
          :required="form.protocol === 'tls'"
        />
        <el-form-item label="客户端认证类型">
          <el-select v-model="form.tls.clientAuthType" placeholder="请选择认证类型">
            <el-option label="no-client-cert" value="no-client-cert" />
            <el-option label="request-client-cert" value="request-client-cert" />
            <el-option label="require-any-client-cert" value="require-any-client-cert" />
            <el-option label="verify-client-cert-if-given" value="verify-client-cert-if-given" />
            <el-option label="require-and-verify-client-cert" value="require-and-verify-client-cert" />
          </el-select>
        </el-form-item>
        <el-form-item label="最低版本">
          <el-select v-model="form.tls.minVersion" placeholder="请选择">
            <el-option label="1.0" value="1.0" />
            <el-option label="1.1" value="1.1" />
            <el-option label="1.2" value="1.2" />
            <el-option label="1.3" value="1.3" />
          </el-select>
        </el-form-item>
        <el-form-item label="最高版本">
          <el-select v-model="form.tls.maxVersion" placeholder="请选择">
            <el-option label="1.0" value="1.0" />
            <el-option label="1.1" value="1.1" />
            <el-option label="1.2" value="1.2" />
            <el-option label="1.3" value="1.3" />
          </el-select>
        </el-form-item>
        <el-form-item label="密码套件">
          <el-checkbox-group v-model="form.tls.cipherSuites">
            <div class="cipher-grid">
              <el-checkbox v-for="cs in TLS_CIPHER_SUITES" :key="cs" :label="cs" :value="cs" />
            </div>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="会话票据">
          <el-switch v-model="form.tls.sessionTickets" />
        </el-form-item>
        <el-form-item label="握手重用">
          <el-switch v-model="form.tls.sessionCache" />
        </el-form-item>
        <el-form-item label="跳过证书验证">
          <el-switch v-model="form.tls.insecureSkipVerify" />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 统计配置卡片 -->
    <el-card style="margin-top: 20px" v-if="form.stats">
      <template #header>
        <span>统计配置</span>
      </template>
      <el-form :model="form" label-width="140px">
        <el-form-item label="启用统计">
          <el-switch v-model="form.stats.enabled" />
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import KeystoreConfig from '@/components/KeystoreConfig.vue'
import { instanceApi, keyStoreApi } from '@/api'
import type { InstanceConfig } from '@/types'

const router = useRouter()

const loading = ref(false)
const keystores = ref<any[]>([])

const TLCP_CIPHER_SUITES = [
  'ECC_SM4_CBC_SM3',
  'ECC_SM4_GCM_SM3',
  'ECDHE_SM4_CBC_SM3',
  'ECDHE_SM4_GCM_SM3'
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
    sessionCache: false,
    insecureSkipVerify: false,
    keystore: undefined,
    keystoreConfig: { type: 'named', name: undefined, params: {} }
  },
  tls: {
    clientAuthType: 'no-client-cert',
    minVersion: '1.2',
    maxVersion: '1.3',
    cipherSuites: [],
    sessionTickets: false,
    sessionCache: false,
    insecureSkipVerify: false,
    keystore: undefined,
    keystoreConfig: { type: 'named', name: undefined, params: {} }
  },
  stats: {
    enabled: false
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

async function create() {
  if (!form.value.name) {
    ElMessage.error('请输入实例名称')
    return
  }

  const protocol = form.value.protocol
  
  // 验证 TLCP keystore 配置（当协议为 tlcp 或 auto 时）
  if (protocol === 'tlcp' || protocol === 'auto') {
    const tlcpConfig = form.value.tlcp.keystoreConfig
    if (!tlcpConfig) {
      ElMessage.error('请配置 TLCP 密钥存储')
      return
    }
    if (tlcpConfig.type === 'named' && !tlcpConfig.name) {
      ElMessage.error('请选择 TLCP 密钥名称')
      return
    }
    if (tlcpConfig.type === 'file') {
      if (!tlcpConfig.params['sign-cert'] || !tlcpConfig.params['sign-key'] ||
          !tlcpConfig.params['enc-cert'] || !tlcpConfig.params['enc-key']) {
        ElMessage.error('请填写完整的 TLCP 密钥文件路径（签名证书、签名密钥、加密证书、加密密钥）')
        return
      }
    }
  }

  // 验证 TLS keystore 配置（当协议为 tls 或 auto 时）
  if (protocol === 'tls' || protocol === 'auto') {
    const tlsConfig = form.value.tls.keystoreConfig
    if (!tlsConfig) {
      ElMessage.error('请配置 TLS 密钥存储')
      return
    }
    if (tlsConfig.type === 'named' && !tlsConfig.name) {
      ElMessage.error('请选择 TLS 密钥名称')
      return
    }
    if (tlsConfig.type === 'file') {
      if (!tlsConfig.params['sign-cert'] || !tlsConfig.params['sign-key']) {
        ElMessage.error('请填写完整的 TLS 密钥文件路径（签名证书、签名密钥）')
        return
      }
    }
  }

  const data: any = { ...form.value }

  if (data.tlcp) {
    data.tlcp.auth = form.value.auth
  }
  if (data.tls) {
    data.tls.auth = form.value.auth
  }

  if (form.value.tlcp.keystoreConfig) {
    const ksConfig = form.value.tlcp.keystoreConfig
    if (ksConfig.type === 'named' && ksConfig.name) {
      if (form.value.protocol === 'tlcp' || form.value.protocol === 'auto') {
        data.tlcp.keystore = { type: 'named', name: ksConfig.name }
      }
    } else if (ksConfig.type === 'file') {
      const hasRequiredFields = ksConfig.params['sign-cert'] && ksConfig.params['sign-key'] &&
        ksConfig.params['enc-cert'] && ksConfig.params['enc-key']
      
      if (hasRequiredFields) {
        if (form.value.protocol === 'tlcp' || form.value.protocol === 'auto') {
          data.tlcp.keystore = {
            type: 'file',
            params: ksConfig.params
          }
        }
      }
    }
  }

  if (form.value.tls.keystoreConfig) {
    const ksConfig = form.value.tls.keystoreConfig
    if (ksConfig.type === 'named' && ksConfig.name) {
      if (form.value.protocol === 'tls' || form.value.protocol === 'auto') {
        data.tls.keystore = { type: 'named', name: ksConfig.name }
      }
    } else if (ksConfig.type === 'file') {
      const hasRequiredFields = ksConfig.params['sign-cert'] && ksConfig.params['sign-key']
      
      if (hasRequiredFields) {
        if (form.value.protocol === 'tls' || form.value.protocol === 'auto') {
          data.tls.keystore = {
            type: 'file',
            params: ksConfig.params
          }
        }
      }
    }
  }

  if (data.tlcp) {
    delete data.tlcp.keystoreConfig
  }
  if (data.tls) {
    delete data.tls.keystoreConfig
  }

  loading.value = true
  try {
    await instanceApi.create(data)
    ElMessage.success('实例创建成功')
    
    // 根据启用状态启动或停止实例
    const instanceName = form.value.name
    if (form.value.enabled) {
      try {
        await instanceApi.start(instanceName)
        ElMessage.success('实例已启动')
      } catch (startErr: any) {
        console.error('启动实例失败:', startErr)
        ElMessage.warning('实例已创建，但启动失败: ' + (startErr.response?.data || startErr.message))
      }
    } else {
      try {
        await instanceApi.stop(instanceName)
        ElMessage.success('实例已停止')
      } catch (stopErr: any) {
        console.error('停止实例失败:', stopErr)
        // 停止失败不影响创建成功，因为默认就是停止状态
      }
    }
    
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
.cipher-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 16px;
}
</style>
