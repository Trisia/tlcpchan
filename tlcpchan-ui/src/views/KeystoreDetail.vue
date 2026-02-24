<template>
  <div class="keystore-detail">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="text-large font-600 mr-3">{{ keystore?.name }}</span>
        <el-tag :type="keystore?.type === 'tlcp' ? 'primary' : 'success'">
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
            <el-tag :type="keystore?.type === 'tlcp' ? 'primary' : 'success'" size="small">
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
          <el-descriptions-item label="创建时间">{{ formatDate(keystore?.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatDate(keystore?.updatedAt) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 证书密钥参数卡片 -->
      <el-card style="margin-top: 20px">
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { keyStoreApi, instanceApi } from '@/api'
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

function formatDate(dateStr: string | undefined): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleString('zh-CN')
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
