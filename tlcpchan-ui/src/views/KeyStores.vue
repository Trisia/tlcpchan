<template>
  <div class="keystores">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>密钥管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            创建密钥
          </el-button>
        </div>
      </template>

      <el-table :data="keystores" v-loading="loading">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.type === CertType.TLCP ? 'primary' : 'success'">{{ row.type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="签名证书/密钥" width="180">
          <template #default="{ row }">
            <span>
              <el-tag v-if="row.hasSignCert" size="small" type="success">证</el-tag>
              <el-tag v-else size="small" type="info">证</el-tag>
              <el-tag v-if="row.hasSignKey" size="small" type="success">钥</el-tag>
              <el-tag v-else size="small" type="info">钥</el-tag>
            </span>
          </template>
        </el-table-column>
        <el-table-column v-if="hasTLCP" label="加密证书/密钥" width="180">
          <template #default="{ row }">
            <span v-if="row.type === CertType.TLCP">
              <el-tag v-if="row.hasEncCert" size="small" type="success">证</el-tag>
              <el-tag v-else size="small" type="info">证</el-tag>
              <el-tag v-if="row.hasEncKey" size="small" type="success">钥</el-tag>
              <el-tag v-else size="small" type="info">钥</el-tag>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="keyParams.algorithm" label="算法" width="100" />
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="showUpdateCertDialog(row)">更新证书</el-button>
            <el-button type="danger" size="small" link @click="remove(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showCreateDialog" title="创建密钥" width="600px">
      <el-form :model="createForm" label-width="120px">
        <el-form-item label="密钥名称" required>
          <el-input v-model="createForm.name" placeholder="my-server-key" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-radio-group v-model="createForm.type">
            <el-radio :value="CertType.TLCP">国密 (TLCP)</el-radio>
            <el-radio :value="CertType.TLS">国际 (TLS)</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="签名证书" required>
          <el-upload
            v-model:file-list="signCertFiles"
            :limit="1"
            :auto-upload="false"
            accept=".crt,.pem"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item label="签名密钥" required>
          <el-upload
            v-model:file-list="signKeyFiles"
            :limit="1"
            :auto-upload="false"
            accept=".key,.pem"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="createForm.type === CertType.TLCP" label="加密证书" required>
          <el-upload
            v-model:file-list="encCertFiles"
            :limit="1"
            :auto-upload="false"
            accept=".crt,.pem"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="createForm.type === CertType.TLCP" label="加密密钥" required>
          <el-upload
            v-model:file-list="encKeyFiles"
            :limit="1"
            :auto-upload="false"
            accept=".key,.pem"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="createKeyStore">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showUpdateCertDialog" title="更新证书" width="500px">
      <el-form label-width="120px">
        <el-form-item label="签名证书">
          <el-upload
            v-model:file-list="updateSignCertFiles"
            :limit="1"
            :auto-upload="false"
            accept=".crt,.pem"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="selectedKeyStore?.type === CertType.TLCP" label="加密证书">
          <el-upload
            v-model:file-list="updateEncCertFiles"
            :limit="1"
            :auto-upload="false"
            accept=".crt,.pem"
          >
            <el-button type="primary">选择文件</el-button>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUpdateCertDialog = false">取消</el-button>
        <el-button type="primary" :loading="updateLoading" @click="updateCertificates">更新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'

const loading = ref(false)
const keystores = ref<any[]>([])

const showCreateDialog = ref(false)
const showUpdateCertDialog = ref(false)
const createLoading = ref(false)
const updateLoading = ref(false)

const createForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls',
})

const signCertFiles = ref<UploadUserFile[]>([])
const signKeyFiles = ref<UploadUserFile[]>([])
const encCertFiles = ref<UploadUserFile[]>([])
const encKeyFiles = ref<UploadUserFile[]>([])

const updateSignCertFiles = ref<UploadUserFile[]>([])
const updateEncCertFiles = ref<UploadUserFile[]>([])
const selectedKeyStore = ref<any>(null)

const hasTLCP = computed(() => keystores.value.some((k) => k.type === CertType.TLCP))

onMounted(() => fetchKeyStores())

async function fetchKeyStores() {
  loading.value = true
  try {
    const result = await keyStoreApi.list()
    keystores.value = result.keystores
  } catch (err) {
    console.error('获取密钥列表失败', err)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

function showUpdateCertDialog(row: any) {
  selectedKeyStore.value = row
  updateSignCertFiles.value = []
  updateEncCertFiles.value = []
  showUpdateCertDialog.value = true
}

async function createKeyStore() {
  if (!createForm.value.name) {
    ElMessage.error('请填写密钥名称')
    return
  }
  if (signCertFiles.value.length === 0 || signKeyFiles.value.length === 0) {
    ElMessage.error('请上传签名证书和密钥')
    return
  }
  if (createForm.value.type === CertType.TLCP && (encCertFiles.value.length === 0 || encKeyFiles.value.length === 0)) {
    ElMessage.error('请上传加密证书和密钥')
    return
  }

  createLoading.value = true
  try {
    const data: any = {
      name: createForm.value.name,
      type: createForm.value.type,
      signCert: signCertFiles.value[0].raw as File,
      signKey: signKeyFiles.value[0].raw as File,
    }
    if (createForm.value.type === CertType.TLCP) {
      data.encCert = encCertFiles.value[0].raw as File
      data.encKey = encKeyFiles.value[0].raw as File
    }

    await keyStoreApi.create(data)
    ElMessage.success('密钥创建成功')
    showCreateDialog.value = false
    resetCreateForm()
    fetchKeyStores()
  } catch (err: any) {
    ElMessage.error(err.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}

async function updateCertificates() {
  if (!selectedKeyStore.value) return
  if (updateSignCertFiles.value.length === 0 && updateEncCertFiles.value.length === 0) {
    ElMessage.error('请至少选择一个证书文件')
    return
  }

  updateLoading.value = true
  try {
    const data: any = {}
    if (updateSignCertFiles.value.length > 0) {
      data.signCert = updateSignCertFiles.value[0].raw as File
    }
    if (updateEncCertFiles.value.length > 0) {
      data.encCert = updateEncCertFiles.value[0].raw as File
    }

    await keyStoreApi.updateCertificates(selectedKeyStore.value.name, data)
    ElMessage.success('证书更新成功')
    showUpdateCertDialog.value = false
    fetchKeyStores()
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败')
  } finally {
    updateLoading.value = false
  }
}

function remove(name: string) {
  ElMessageBox.confirm('确定要删除此密钥吗？', '确认删除', { type: 'warning' })
    .then(async () => {
      try {
        await keyStoreApi.delete(name)
        ElMessage.success('密钥已删除')
        fetchKeyStores()
      } catch (err) {
        console.error('删除失败', err)
      }
    })
    .catch(() => {})
}

function resetCreateForm() {
  createForm.value = {
    name: '',
    type: CertType.TLCP,
  }
  signCertFiles.value = []
  signKeyFiles.value = []
  encCertFiles.value = []
  encKeyFiles.value = []
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
