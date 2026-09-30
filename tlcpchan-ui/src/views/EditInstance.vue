<template>
  <div class="edit-instance">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">编辑实例: {{ form.name }}</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px">
      <el-form :model="form" label-width="120px" v-loading="loading">
        <el-form-item label="实例名称">
          <el-input v-model="form.name" disabled />
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
        <el-form-item label="监听地址" required>
          <el-input v-model="form.listen" />
        </el-form-item>
        <el-form-item label="目标地址" required>
          <el-input v-model="form.target" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="save" :loading="loading">保存</el-button>
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
        <!-- 分组一：PKI 配置（X.509 证书身份） -->
        <el-divider content-position="left">PKI 配置</el-divider>
        <KeystoreConfig
          v-model="form.tlcp.keystoreConfig"
          :keystores="keystores"
          :is-tlcp="true"
          :required="tlcpKeystoreRequired"
        />
        <!-- 分组二：IBC 配置（SM9 标识身份，可选，与证书身份并列） -->
        <el-divider content-position="left">IBC 配置</el-divider>
        <el-alert
          title="IBC(SM9) 身份与证书身份相互独立、可同时配置以实现同端口混合协商；未配置时 IBC/IBSDH 套件不可用。"
          type="info"
          :closable="false"
          style="margin-bottom: 16px"
        />
        <KeystoreConfig
          v-model="form.tlcp.ibcKeystoreConfig"
          :keystores="keystores"
          :is-tlcp="true"
          ibc
        />
        <!-- 分组三：TLCP 协议参数（含 IBC/IBSDH 套件勾选） -->
        <el-divider content-position="left">TLCP 协议配置</el-divider>
        <el-form-item label="客户端认证类型">
          <el-select v-model="form.tlcp.clientAuthType" placeholder="请选择认证类型">
            <el-option label="不要求证书（no-client-cert）" value="no-client-cert" />
            <el-option label="请求证书（request-client-cert）" value="request-client-cert" />
            <el-option label="要求证书（require-any-client-cert）" value="require-any-client-cert" />
            <el-option label="提供则验证（verify-client-cert-if-given）" value="verify-client-cert-if-given" />
            <el-option label="要求并验证（require-and-verify-client-cert）" value="require-and-verify-client-cert" />
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
              <el-tooltip
                v-for="cs in TLCP_CIPHER_SUITES"
                :key="cs"
                :disabled="!cipherDisabledReason(cs)"
                :content="cipherDisabledReason(cs)"
                placement="top"
              >
                <span>
                  <el-checkbox :label="cs" :value="cs" :disabled="!!cipherDisabledReason(cs)" />
                </span>
              </el-tooltip>
            </div>
          </el-checkbox-group>
          <div v-if="!hasTlcpKeystore" class="cipher-hint">ECDHE 套件需先配置 TLCP 证书 keystore</div>
          <!-- 证书身份已配齐时，服务端还需把客户端认证类型设为要求客户端证书，ECDHE 才可选 -->
          <div v-if="hasTlcpKeystore && !ecdheClientAuthOk" class="cipher-hint">
            ECDHE 套件要求认证客户端身份，请将客户端认证类型设为「要求证书」或「要求并验证」
          </div>
          <div v-if="!hasIbcKeystore" class="cipher-hint">IBC/IBSDH 套件需先配置 IBC 身份 keystore</div>
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
          :required="tlsKeystoreRequired"
        />
        <el-form-item label="客户端认证类型">
          <el-select v-model="form.tls.clientAuthType" placeholder="请选择认证类型">
            <el-option label="不要求证书（no-client-cert）" value="no-client-cert" />
            <el-option label="请求证书（request-client-cert）" value="request-client-cert" />
            <el-option label="要求证书（require-any-client-cert）" value="require-any-client-cert" />
            <el-option label="提供则验证（verify-client-cert-if-given）" value="verify-client-cert-if-given" />
            <el-option label="要求并验证（require-and-verify-client-cert）" value="require-and-verify-client-cert" />
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
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import KeystoreConfig from '@/components/KeystoreConfig.vue'
import { instanceApi, keyStoreApi } from '@/api'
import { isOneWayClientAuth } from '@/constants/clientAuth'
import { TLCP_CIPHER_SUITES, TLS_CIPHER_SUITES, isIBCSuite, isClientAuthAllowedForECDHE, cipherSuiteDisabledReason } from '@/constants/cipherSuite'
import type { InstanceConfig, KeystoreConfigUI } from '@/types'

const route = useRoute()
const router = useRouter()

const instanceName = route.params.name as string

const loading = ref(false)
const keystores = ref<any[]>([])

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
    keystoreConfig: { type: 'named', name: undefined, params: {} },
    ibcKeystore: undefined,
    ibcKeystoreConfig: { type: 'named', name: undefined, params: {} }
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

onMounted(async () => {
  await Promise.all([loadInstance(), loadKeystores()])
})

async function loadInstance() {
  loading.value = true
  try {
    const data = await instanceApi.get(instanceName)
    form.value = data.config
    
    // 解析 TLCP keystore 配置
    if (form.value.tlcp?.keystore) {
      form.value.tlcp.keystoreConfig = parseKeystoreConfig(form.value.tlcp.keystore)
    } else {
      form.value.tlcp.keystoreConfig = { type: 'named', name: undefined, params: {} }
    }

    // 解析 TLCP IBC 身份 keystore 配置
    if (form.value.tlcp?.ibcKeystore) {
      form.value.tlcp.ibcKeystoreConfig = parseKeystoreConfig(form.value.tlcp.ibcKeystore)
    } else {
      form.value.tlcp.ibcKeystoreConfig = { type: 'named', name: undefined, params: {} }
    }
    
    // 解析 TLS keystore 配置
    if (form.value.tls?.keystore) {
      form.value.tls.keystoreConfig = parseKeystoreConfig(form.value.tls.keystore)
    } else {
      form.value.tls.keystoreConfig = { type: 'named', name: undefined, params: {} }
    }

    // 设置统计配置默认值
    if (!form.value.stats) {
      form.value.stats = {
        enabled: false
      }
    }

    // 兜底清理：历史配置可能残留“已勾选 ECDHE 但前置条件不满足”的组合，
    // 该组合在界面上会呈现为置灰且勾选，用户无法主动取消
    dropInvalidEcdheSuites()
  } catch (err) {
    console.error('加载实例失败:', err)
    ElMessage.error('加载实例失败')
  } finally {
    loading.value = false
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

// 解析 keystore 配置为 UI 格式
function parseKeystoreConfig(keystore: any): KeystoreConfigUI {
  if (keystore.type === 'named') {
    return {
      type: 'named',
      name: keystore.name,
      params: {}
    }
  } else if (keystore.type === 'file' || keystore.type === 'ibc-file') {
    return {
      type: keystore.type,
      name: undefined,
      params: keystore.params || {}
    }
  }
  return { type: 'named', name: undefined, params: {} }
}

/**
 * 判断套件是否为 ECDHE 套件
 *
 * ECDHE 套件要求本端提供证书身份（TLCP 证书 keystore）参与密钥交换，
 * 且服务端必须对客户端做身份认证（客户端认证类型需为“要求证书/要求并验证”），
 * 任一条件不满足时该套件不可选。
 * @param suite TLCP 密码套件名
 */
function isEcdheSuite(suite: string): boolean {
  return suite.startsWith('ECDHE_')
}

// 是否已配置可用的 IBC 身份 keystore（named 需选中名称，ibc-file 需填写标识与公共参数）
const hasIbcKeystore = computed(() => {
  const cfg = form.value.tlcp?.ibcKeystoreConfig
  if (!cfg) return false
  if (cfg.type === 'named') return !!cfg.name
  if (cfg.type === 'ibc-file') return !!(cfg.params['identity'] && cfg.params['params'])
  return false
})

// 是否已配置完整的 TLCP 证书 keystore（决定 ECDHE 套件是否可选）
const hasTlcpKeystore = computed(() => {
  const cfg = form.value.tlcp?.keystoreConfig
  if (!cfg) return false
  if (cfg.type === 'named') return !!cfg.name
  if (cfg.type === 'file') {
    return !!(cfg.params['sign-cert'] && cfg.params['sign-key'] &&
      cfg.params['enc-cert'] && cfg.params['enc-key'])
  }
  return false
})

// 是否为客户端角色（客户端代理的证书身份用于向服务端证明自身，而非验证对端）
const isClientRole = computed(() => form.value.type === 'client' || form.value.type === 'http-client')

// 是否为服务端角色（服务端才会要求客户端出示证书，因此 ECDHE 的认证类型约束只作用于服务端）
const isServerRole = computed(() => !isClientRole.value)

// ECDHE 的客户端认证类型前置条件是否满足（仅服务端角色要求；客户端代理由对端决定是否要求本端证书）
const ecdheClientAuthOk = computed(() => !isServerRole.value || isClientAuthAllowedForECDHE(form.value.tlcp?.clientAuthType))

/**
 * 判断指定协议下是否处于单向认证（本端作为客户端且对端不强制本端出示证书）
 * @param protocol 协议类型，'tlcp' 或 'tls'
 * @returns true 表示单向认证，此时客户端代理无需配置 keystore
 */
function isOneWayAuth(protocol: 'tlcp' | 'tls'): boolean {
  if (!isClientRole.value) return false
  const authType = protocol === 'tlcp' ? form.value.tlcp.clientAuthType : form.value.tls.clientAuthType
  return isOneWayClientAuth(authType)
}

// TLCP 证书 keystore 是否必填：协议指定 TLCP、未配置 IBC 身份、且非客户端单向认证
const tlcpKeystoreRequired = computed(() => {
  if (form.value.protocol !== 'tlcp') return false
  if (hasIbcKeystore.value) return false
  return !isOneWayAuth('tlcp')
})

// TLS keystore 是否必填：协议指定 TLS 且非客户端单向认证
const tlsKeystoreRequired = computed(() => form.value.protocol === 'tls' && !isOneWayAuth('tls'))

/**
 * 返回密码套件不可选的原因
 * @param suite TLCP 密码套件名
 * @returns 不可选原因文本；套件可选时返回空字符串
 */
function cipherDisabledReason(suite: string): string {
  const ibcReason = cipherSuiteDisabledReason(suite, hasIbcKeystore.value)
  if (ibcReason) return ibcReason
  if (isEcdheSuite(suite)) {
    if (!hasTlcpKeystore.value) return '需先配置 TLCP 证书 keystore'
    if (!ecdheClientAuthOk.value) return 'ECDHE 需认证客户端身份，请将客户端认证类型设为「要求证书」或「要求并验证」'
  }
  return ''
}

// 未配置 IBC 身份时移除已勾选的 IBC/IBSDH 套件，避免提交出无效组合
watch(hasIbcKeystore, (ok) => {
  if (!ok && form.value.tlcp?.cipherSuites) {
    form.value.tlcp.cipherSuites = form.value.tlcp.cipherSuites.filter((cs) => !isIBCSuite(cs))
  }
})

/**
 * 移除当前已不满足前置条件的 ECDHE 套件
 *
 * ECDHE 的前置条件为“证书 keystore 已配齐”与“服务端客户端认证类型合格”。
 * 条件不满足时复选框会置灰，而置灰的复选框无法被点击取消，
 * 因此必须在条件变化或配置加载完成时主动清理，否则用户无法移除该项。
 */
function dropInvalidEcdheSuites() {
  if (!form.value.tlcp?.cipherSuites) return
  if (hasTlcpKeystore.value && ecdheClientAuthOk.value) return
  form.value.tlcp.cipherSuites = form.value.tlcp.cipherSuites.filter((cs) => !isEcdheSuite(cs))
}

// 证书 keystore 或服务端客户端认证类型不再满足要求时，移除已勾选的 ECDHE 套件
watch([hasTlcpKeystore, ecdheClientAuthOk], dropInvalidEcdheSuites)

async function save() {
  const protocol = form.value.protocol

  // 验证 TLCP 身份配置（当协议为 tlcp 或 auto 时）
  // 客户端单向认证无需 keystore；其余场景证书 keystore 与 IBC 身份 keystore 至少配置其一
  if (protocol === 'tlcp' || protocol === 'auto') {
    const tlcpConfig = form.value.tlcp?.keystoreConfig
    if (!hasIbcKeystore.value && !isOneWayAuth('tlcp')) {
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

    // IBC 直接指定文件时，标识与 KGC 公共参数必须成对填写
    const ibcConfig = form.value.tlcp?.ibcKeystoreConfig
    if (ibcConfig && ibcConfig.type === 'ibc-file') {
      const hasIdentity = !!ibcConfig.params['identity']
      const hasParams = !!ibcConfig.params['params']
      if (hasIdentity !== hasParams) {
        ElMessage.error('请同时填写 IBC 标识文件路径与 KGC 公共参数路径')
        return
      }
    }
  }

  // 验证 TLS keystore 配置（当协议为 tls 或 auto 时，客户端单向认证无需 keystore）
  if (protocol === 'tls' || protocol === 'auto') {
    const tlsConfig = form.value.tls?.keystoreConfig
    if (!isOneWayAuth('tls')) {
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
  }

  loading.value = true
  try {
    const data: any = { ...form.value }

    // 确保 tlcp 对象存在
    if (!data.tlcp) {
      data.tlcp = {
        clientAuthType: 'no-client-cert',
        minVersion: '1.1',
        maxVersion: '1.1',
        cipherSuites: [],
        sessionCache: false,
        insecureSkipVerify: false
      }
    }

    // 确保 tls 对象存在
    if (!data.tls) {
      data.tls = {
        clientAuthType: 'no-client-cert',
        minVersion: '1.2',
        maxVersion: '1.3',
        cipherSuites: [],
        sessionTickets: false,
        sessionCache: false,
        insecureSkipVerify: false
      }
    }

    // 处理 TLCP keystore 配置
    if (form.value.tlcp?.keystoreConfig) {
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

    // 处理 TLCP IBC 身份 keystore 配置（可选，与证书 keystore 并列）
    // 先清除既有引用，避免用户移除 IBC 身份配置后仍提交旧值
    delete data.tlcp.ibcKeystore
    if (form.value.tlcp?.ibcKeystoreConfig) {
      const ibcConfig = form.value.tlcp.ibcKeystoreConfig
      if (form.value.protocol === 'tlcp' || form.value.protocol === 'auto') {
        if (ibcConfig.type === 'named' && ibcConfig.name) {
          data.tlcp.ibcKeystore = { type: 'named', name: ibcConfig.name }
        } else if (ibcConfig.type === 'ibc-file' && ibcConfig.params['identity'] && ibcConfig.params['params']) {
          data.tlcp.ibcKeystore = {
            type: 'ibc-file',
            params: ibcConfig.params
          }
        }
      }
    }

    // 处理 TLS keystore 配置
    if (form.value.tls?.keystoreConfig) {
      const ksConfig = form.value.tls.keystoreConfig
      if (ksConfig.type === 'named' && ksConfig.name) {
        if (form.value.protocol === 'tls' || form.value.protocol === 'auto') {
          data.tls.keystore = { type: 'named', name: ksConfig.name }
        }
      } else if (ksConfig.type === 'file') {
        const hasRequiredFields = ksConfig.params['sign-cert'] &&ksConfig.params['sign-key']
        
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

    // 删除 keystoreConfig / ibcKeystoreConfig 字段（前端使用字段）
    if (data.tlcp) {
      delete data.tlcp.keystoreConfig
      delete data.tlcp.ibcKeystoreConfig
    }
    if (data.tls) {
      delete data.tls.keystoreConfig
    }

    // 未配置 IBC 身份时不提交 IBC/IBSDH 套件（手工编辑过的旧配置可能残留）
    if (!hasIbcKeystore.value && data.tlcp?.cipherSuites) {
      data.tlcp.cipherSuites = data.tlcp.cipherSuites.filter((cs: string) => !isIBCSuite(cs))
    }

    // 未配置任何 TLCP 证书 keystore 时不提交 ECDHE 套件（手工编辑过的旧配置可能残留）
    // 注意：以提交数据中的 keystore 引用为准，硬件类 keystore（skf/sdf）无法在前端表单中表达，不能误删
    if (!data.tlcp.keystore && data.tlcp?.cipherSuites) {
      data.tlcp.cipherSuites = data.tlcp.cipherSuites.filter((cs: string) => !isEcdheSuite(cs))
    }

    await instanceApi.update(instanceName, data)
    ElMessage.success('实例配置已保存')
    
    // 重载实例配置
    try {
      await instanceApi.reload(instanceName)
      ElMessage.success('实例配置已重载')
    } catch (reloadErr: any) {
      console.error('重载实例失败:', reloadErr)
      ElMessage.warning('实例配置已保存，但重载失败: ' + (reloadErr.response?.data || reloadErr.message))
    }
    
    router.push(`/instances/${instanceName}`)
  } catch (err: any) {
    console.error('保存失败:', err)
    ElMessage.error('保存失败: ' + (err.response?.data || err.message))
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
.cipher-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #e6a23c;
}
</style>