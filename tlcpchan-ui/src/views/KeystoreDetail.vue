<template>
  <div class="keystore-detail">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">{{ keystore?.name }}</span>
        <el-tag :type="typeTagType(keystore?.type)">
          {{ keystore?.type?.toUpperCase() }}
        </el-tag>
      </template>
    </el-page-header>

    <div style="margin-top: 20px">
      <!-- 基本信息卡片 -->
      <el-card>
        <template #header>
          <span>基本信息</span>
        </template>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="名称">{{ keystore?.name }}</el-descriptions-item>
          <el-descriptions-item label="类型">
            <el-tag :type="typeTagType(keystore?.type)" size="small">
              {{ keystore?.type?.toUpperCase() }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="加载器类型">
            <el-tag size="small">{{ keystore?.loaderType }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="保护状态">
            <el-tag :type="keystore?.protected ? 'danger' : 'info'" size="small">
              {{ keystore?.protected ? '受保护' : '不受保护' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDateTime(keystore?.createdAt, '') }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDateTime(keystore?.updatedAt, '') }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- IBC 身份元信息卡片（仅 type=ibc） -->
      <el-card v-if="keystore?.type === 'ibc'" style="margin-top: 20px">
        <template #header>
          <span>IBC 身份信息</span>
        </template>

        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="标识">{{ keystore?.ibc?.identity || '-' }}</el-descriptions-item>
          <el-descriptions-item label="KGC 区域">{{ keystore?.ibc?.districtName || '-' }}</el-descriptions-item>
          <el-descriptions-item label="KGC 序号">{{ keystore?.ibc?.districtName ? keystore?.ibc?.districtSerial : '-' }}</el-descriptions-item>
          <el-descriptions-item label="KGC 公共参数">
            <el-tag :type="keystore?.ibc?.hasParams ? 'success' : 'info'" size="small">
              {{ keystore?.ibc?.hasParams ? '已提供' : '未提供' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="生效时间">
            {{ formatDateTime(keystore?.ibc?.notBefore, '不限') }}
          </el-descriptions-item>
          <el-descriptions-item label="失效时间">
            {{ formatDateTime(keystore?.ibc?.notAfter, '不限') }}
          </el-descriptions-item>
          <el-descriptions-item label="签名私钥 (hid=0x01)">
            <el-tag :type="keystore?.ibc?.hasSignKey ? 'success' : 'info'" size="small">
              {{ keystore?.ibc?.hasSignKey ? '已提供' : '未提供' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="加密私钥 (hid=0x03)">
            <el-tag :type="keystore?.ibc?.hasEncryptKey ? 'success' : 'info'" size="small">
              {{ keystore?.ibc?.hasEncryptKey ? '已提供' : '未提供' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="密钥交换私钥 (hid=0x02)">
            <el-tag :type="keystore?.ibc?.hasKeyExchangeKey ? 'success' : 'info'" size="small">
              {{ keystore?.ibc?.hasKeyExchangeKey ? '已提供' : '未提供' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <el-alert
          title="签名私钥用于签名/验签（hid=0x01），加密私钥用于解密预主密钥（hid=0x03），密钥交换私钥用于 SM9 密钥交换（hid=0x02），三者不可混用。"
          type="info"
          :closable="false"
          style="margin-top: 16px"
        />

        <el-descriptions :column="1" border size="small" style="margin-top: 16px">
          <el-descriptions-item label="标识文件">{{ keystore?.params?.['identity'] || '-' }}</el-descriptions-item>
          <el-descriptions-item label="公共参数文件">{{ keystore?.params?.['params'] || '-' }}</el-descriptions-item>
          <el-descriptions-item label="签名私钥文件">{{ keystore?.params?.['sign-key'] || '-' }}</el-descriptions-item>
          <el-descriptions-item label="加密私钥文件">{{ keystore?.params?.['enc-key'] || '-' }}</el-descriptions-item>
          <el-descriptions-item label="密钥交换私钥文件">{{ keystore?.params?.['kex-key'] || '-' }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- IBC 材料替换卡片（仅 type=ibc 且未受保护） -->
      <el-card v-if="keystore?.type === 'ibc' && !keystore?.protected" style="margin-top: 20px">
        <template #header>
          <span>替换 IBC 材料</span>
        </template>

        <el-alert
          title="所有文件均为可选，仅上传需要替换的材料；替换后需重新加载关联实例才能生效。"
          type="info"
          :closable="false"
          style="margin-bottom: 16px"
        />

        <el-form label-width="160px">
          <el-form-item label="标识文件">
            <el-upload v-model:file-list="ibcIdentityFiles" :limit="1" :auto-upload="false" accept=".txt,.der,.pem">
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item label="KGC 公共参数">
            <el-upload v-model:file-list="ibcParamsFiles" :limit="1" :auto-upload="false" accept=".pem,.der,.ibcparams">
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item label="签名私钥 (hid=0x01)">
            <el-upload v-model:file-list="ibcSignKeyFiles" :limit="1" :auto-upload="false" accept=".key,.pem">
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item label="加密私钥 (hid=0x03)">
            <el-upload v-model:file-list="ibcEncKeyFiles" :limit="1" :auto-upload="false" accept=".key,.pem">
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          <el-form-item label="密钥交换私钥 (hid=0x02)">
            <el-upload v-model:file-list="ibcKexKeyFiles" :limit="1" :auto-upload="false" accept=".key,.pem">
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
        </el-form>

        <div style="margin-top: 16px">
          <el-button type="primary" @click="handleReplaceIBC" :loading="replacing">
            替换材料
          </el-button>
        </div>
      </el-card>

      <!-- 证书密钥参数卡片（IBC 类型不适用） -->
      <el-card v-if="keystore?.type !== 'ibc'" style="margin-top: 20px">
        <template #header>
          <span>证书密钥参数</span>
        </template>
        
        <el-alert
          v-if="keystore?.protected"
          title="受保护的 keystore 不允许修改"
          type="warning"
          :closable="false"
          style="margin-bottom: 16px"
        />
        
        <el-alert
          v-if="keystore?.loaderType !== 'file'"
          title="只有文件类型的 keystore 支持编辑参数"
          type="info"
          :closable="false"
          style="margin-bottom: 16px"
        />

        <el-form
          ref="formRef"
          :model="editableParams"
          label-width="120px"
          :disabled="isEditDisabled"
        >
          <template v-if="keystore?.type === 'tlcp'">
            <el-form-item label="签名证书路径">
              <el-input v-model="editableParams['sign-cert']" placeholder="sign.crt" />
            </el-form-item>
            <el-form-item label="签名密钥路径">
              <el-input v-model="editableParams['sign-key']" placeholder="sign.key" />
            </el-form-item>
            <el-form-item label="加密证书路径">
              <el-input v-model="editableParams['enc-cert']" placeholder="enc.crt" />
            </el-form-item>
            <el-form-item label="加密密钥路径">
              <el-input v-model="editableParams['enc-key']" placeholder="enc.key" />
            </el-form-item>
          </template>
          
          <template v-if="keystore?.type === 'tls'">
            <el-form-item label="证书路径">
              <el-input v-model="editableParams['cert']" placeholder="server.crt" />
            </el-form-item>
            <el-form-item label="密钥路径">
              <el-input v-model="editableParams['key']" placeholder="server.key" />
            </el-form-item>
          </template>
        </el-form>

        <div v-if="!isEditDisabled" style="margin-top: 16px">
          <el-button type="primary" @click="handleSave" :loading="saving">
            保存修改
          </el-button>
        </div>
      </el-card>

      <!-- 关联实例卡片 -->
      <el-card style="margin-top: 20px">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <span>关联实例</span>
            <el-button
              v-if="runningInstances.length > 0"
              type="warning"
              size="small"
              @click="handleReloadAllInstances"
              :loading="reloading"
            >
              重新加载所有关联实例 ({{ runningInstances.length }})
            </el-button>
          </div>
        </template>

        <el-empty v-if="!relatedInstances || relatedInstances.length === 0" description="暂无关联实例" />

        <el-table v-else :data="relatedInstances" v-loading="instancesLoading">
          <el-table-column prop="name" label="实例名称" />
          <el-table-column prop="protocol" label="协议类型" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="row.protocol === 'tlcp' ? 'primary' : 'success'">
                {{ row.protocol?.toUpperCase() }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="statusType(row.status)">
                {{ statusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <el-alert
          v-if="showReloadWarning"
          !title="修改后需要重新加载关联实例才能生效"
          type="warning"
          :closable="false"
          style="margin-top: 16px"
        />
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { keyStoreApi, instanceApi } from '@/api'
import { formatDateTime } from '@/constants/datetime'
import type { KeyStoreInfo, KeystoreInstance } from '@/types'

const route = useRoute()
const router = useRouter()

const keystore = ref<KeyStoreInfo | null>(null)
const relatedInstances = ref<KeystoreInstance[]>([])
const editableParams = ref<Record<string, string>>({})
const saving = ref(false)
const reloading = ref(false)
const instancesLoading = ref(false)
const showReloadWarning = ref(false)
const replacing = ref(false)

// IBC 材料替换文件
const ibcIdentityFiles = ref<UploadUserFile[]>([])
const ibcParamsFiles = ref<UploadUserFile[]>([])
const ibcSignKeyFiles = ref<UploadUserFile[]>([])
const ibcEncKeyFiles = ref<UploadUserFile[]>([])
const ibcKexKeyFiles = ref<UploadUserFile[]>([])

const name = computed(() => route.params.name as string)

const isEditDisabled = computed(() => {
  return keystore.value?.protected || keystore.value?.loaderType !== 'file'
})

const runningInstances = computed(() => {
  return relatedInstances.value.filter((inst) => inst.status === 'running')
})

onMounted(() => {
  fetchKeystore()
  fetchRelatedInstances()
})

async function fetchKeystore() {
  try {
    keystore.value = await keyStoreApi.get(name.value)
    editableParams.value = { ...keystore.value?.params }
  } catch (err: any) {
    ElMessage.error(`获取 keystore 失败: ${err.message || '未知错误'}`)
    router.back()
  }
}

async function fetchRelatedInstances() {
  instancesLoading.value = true
  try {
    relatedInstances.value = await keyStoreApi.getInstances(name.value)
  } catch (err: any) {
    console.error('获取关联实例失败:', err)
  } finally {
    instancesLoading.value = false
  }
}

/**
 * keystore 类型对应的标签颜色
 * @param type keystore 类型（tlcp / tls / ibc）
 */
function typeTagType(type: string | undefined): 'primary' | 'success' | 'warning' {
  if (type === 'tlcp') return 'primary'
  if (type === 'ibc') return 'warning'
  return 'success'
}

function statusType(status: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    running: 'success',
    stopped: 'info',
    error: 'danger',
    created: 'warning'
  }
  return map[status] || ''
}

function statusText(status: string): string {
  const map: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    error: '错误',
    created: '已创建'
  }
  return map[status] || status
}

async function handleSave() {
  if (saving.value) return

  const changedParams: Record<string, string> = {}
  for (const key in editableParams.value) {
    const oldValue = keystore.value?.params[key] || ''
    const newValue = editableParams.value[key] || ''
    if (newValue !== oldValue) {
      changedParams[key] = newValue
    }
  }

  if (Object.keys(changedParams).length === 0) {
    ElMessage.info('没有修改的内容')
    return
  }

  saving.value = true
  try {
    await keyStoreApi.update(name.value, { params: changedParams })
    ElMessage.success('保存成功')
    
    await fetchKeystore()
    await fetchRelatedInstances()
    
    if (runningInstances.value.length > 0) {
      showReloadWarning.value = true
    }
  } catch (err: any) {
    ElMessage.error(`保存失败: ${err.response?.data || err.message || '未知错误'}`)
  } finally {
    saving.value = false
  }
}

/**
 * 替换 IBC 身份材料（仅上传选中的文件）
 */
async function handleReplaceIBC() {
  const data: any = {}
  if (ibcIdentityFiles.value.length > 0) data.identity = ibcIdentityFiles.value[0]?.raw as File
  if (ibcParamsFiles.value.length > 0) data.params = ibcParamsFiles.value[0]?.raw as File
  if (ibcSignKeyFiles.value.length > 0) data.signKey = ibcSignKeyFiles.value[0]?.raw as File
  if (ibcEncKeyFiles.value.length > 0) data.encKey = ibcEncKeyFiles.value[0]?.raw as File
  if (ibcKexKeyFiles.value.length > 0) data.kexKey = ibcKexKeyFiles.value[0]?.raw as File

  if (Object.keys(data).length === 0) {
    ElMessage.info('请至少选择一个要替换的材料文件')
    return
  }

  replacing.value = true
  try {
    await keyStoreApi.updateCertificates(name.value, data)
    ElMessage.success('IBC 材料替换成功')

    ibcIdentityFiles.value = []
    ibcParamsFiles.value = []
    ibcSignKeyFiles.value = []
    ibcEncKeyFiles.value = []
    ibcKexKeyFiles.value = []

    await fetchKeystore()
    await fetchRelatedInstances()

    if (runningInstances.value.length > 0) {
      showReloadWarning.value = true
    }
  } catch (err: any) {
    ElMessage.error(`替换失败: ${err.response?.data || err.message || '未知错误'}`)
  } finally {
    replacing.value = false
  }
}

async function handleReloadAllInstances() {
  if (reloading.value) return

  try {
    await ElMessageBox.confirm(
      '确定要重新加载所有关联的运行中实例吗？',
      '确认重载',
      { type: 'warning' }
    )
  } catch {
    return
  }

  reloading.value = true
  const failedInstances: string[] = []

  for (const inst of runningInstances.value) {
    try {
      await instanceApi.reload(inst.name)
    } catch (err: any) {
      console.error(`重载实例 ${inst.name} 失败:`, err)
      failedInstances.push(inst.name)
    }
  }

  if (failedInstances.length === 0) {
    ElMessage.success('所有实例重载成功')
  } else {
    ElMessage.warning(`部分实例重载失败: ${failedInstances.join(', ')}`)
  }

  await fetchRelatedInstances()
  showReloadWarning.value = false
  reloading.value = false
}
</script>

<style scoped>
.text-large {
  font-size: 18px;
}

.mr-3 {
  margin-right: 12px;
}

.font-600 {
  font-weight: 600;
}
</style>
