<template>
  <div class="keystores">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>密钥管理</span>
          <div>
            <el-button type="success" @click="showGenerateDialog = true">
              <el-icon><MagicStick /></el-icon>
              生成密钥
            </el-button>
            <el-button type="primary" @click="showCreateDialog = true">
              <el-icon><Plus /></el-icon>
              创建密钥
            </el-button>
          </div>
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
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="success" size="small" link @click="showExportCSRDialog(row)">导出 CSR</el-button>
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

    <el-dialog v-model="showGenerateDialog" title="生成密钥" width="700px">
      <el-form :model="generateForm" label-width="140px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="密钥名称" required>
              <el-input v-model="generateForm.name" placeholder="my-server-key" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="类型" required>
              <el-radio-group v-model="generateForm.type">
                <el-radio :value="CertType.TLCP">国密 (TLCP)</el-radio>
                <el-radio :value="CertType.TLS">国际 (TLS)</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">证书主体 (DN)</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="国家 (C)">
              <el-input v-model="generateForm.certConfig.country" placeholder="CN" maxlength="2" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="省/州 (ST)">
              <el-input v-model="generateForm.certConfig.stateOrProvince" placeholder="Beijing" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="地区 (L)">
              <el-input v-model="generateForm.certConfig.locality" placeholder="Haidian" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="组织 (O)">
              <el-input v-model="generateForm.certConfig.org" placeholder="Example Org" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="组织单位 (OU)">
              <el-input v-model="generateForm.certConfig.orgUnit" placeholder="IT" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="通用名称 (CN)" required>
              <el-input v-model="generateForm.certConfig.commonName" placeholder="example.com" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱地址">
              <el-input v-model="generateForm.certConfig.emailAddress" placeholder="admin@example.com" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">有效期</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="有效期(年)">
              <el-input-number v-model="generateForm.certConfig.years" :min="1" :max="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="或 有效期(天)">
              <el-input-number v-model="generateForm.certConfig.days" :min="1" :max="36500" />
            </el-form-item>
          </el-col>
        </el-row>

        <template v-if="generateForm.type === CertType.TLS">
          <el-divider content-position="left">密钥选项 (仅 TLS)</el-divider>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="密钥算法">
                <el-select v-model="generateForm.certConfig.keyAlgorithm">
                  <el-option label="ECDSA" value="ecdsa" />
                  <el-option label="RSA" value="rsa" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="密钥位数 (仅 RSA)">
                <el-select v-model="generateForm.certConfig.keyBits">
                  <el-option :label="2048" :value="2048" />
                  <el-option :label="4096" :value="4096" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <el-divider content-position="left">主题备用名称 (SAN)</el-divider>
        <el-form-item label="DNS 名称">
          <el-select
            v-model="generateForm.certConfig.dnsNames"
            multiple
            filterable
            allow-create
            placeholder="添加 DNS 名称，如 example.com"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-select
            v-model="generateForm.certConfig.ipAddresses"
            multiple
            filterable
            allow-create
            placeholder="添加 IP 地址，如 192.168.1.1"
            style="width: 100%"
          />
        </el-form-item>

        <el-form-item label="受保护">
          <el-switch v-model="generateForm.protected" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGenerateDialog = false">取消</el-button>
        <el-button type="primary" :loading="generateLoading" @click="generateKeyStore">生成</el-button>
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

    <el-dialog v-model="showExportCSRDialog" title="导出证书请求 (CSR)" width="600px">
      <el-form :model="exportCSRForm" label-width="140px">
        <el-form-item v-if="selectedKeyStore?.type === CertType.TLCP" label="密钥类型" required>
          <el-radio-group v-model="exportCSRForm.keyType">
            <el-radio value="sign">签名密钥</el-radio>
            <el-radio value="enc">加密密钥</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-divider content-position="left">证书主体 (DN)</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="通用名称 (CN)" required>
              <el-input v-model="exportCSRForm.csrParams.commonName" placeholder="example.com" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="组织 (O)">
              <el-input v-model="exportCSRForm.csrParams.org" placeholder="Example Org" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="国家 (C)">
              <el-input v-model="exportCSRForm.csrParams.country" placeholder="CN" maxlength="2" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="省/州 (ST)">
              <el-input v-model="exportCSRForm.csrParams.stateOrProvince" placeholder="Beijing" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="地区 (L)">
              <el-input v-model="exportCSRForm.csrParams.locality" placeholder="Beijing" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="组织单位 (OU)">
              <el-input v-model="exportCSRForm.csrParams.orgUnit" placeholder="IT" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱地址">
              <el-input v-model="exportCSRForm.csrParams.emailAddress" placeholder="admin@example.com" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">主题备用名称 (SAN)</el-divider>
        <el-form-item label="DNS 名称">
          <el-select
            v-model="exportCSRForm.csrParams.dnsNames"
            multiple
            filterable
            allow-create
            placeholder="添加 DNS 名称，如 example.com"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-select
            v-model="exportCSRForm.csrParams.ipAddresses"
            multiple
            filterable
            allow-create
            placeholder="添加 IP 地址，如 192.168.1.1"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showExportCSRDialog = false">取消</el-button>
        <el-button type="primary" :loading="exportCSRLoading" @click="exportCSR">导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { Plus, MagicStick } from '@element-plus/icons-vue'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'

const loading = ref(false)
const keystores = ref<any[]>([])

const showCreateDialog = ref(false)
const showGenerateDialog = ref(false)
const showUpdateCertDialog = ref(false)
const showExportCSRDialog = ref(false)
const createLoading = ref(false)
const generateLoading = ref(false)
const updateLoading = ref(false)
const exportCSRLoading = ref(false)

const createForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls',
})

const generateForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls',
  protected: false,
  certConfig: {
    commonName: '',
    country: '',
    stateOrProvince: '',
    locality: '',
    org: '',
    orgUnit: '',
    emailAddress: '',
    years: 1,
    days: 0,
    keyAlgorithm: 'ecdsa' as string,
    keyBits: 2048 as number,
    dnsNames: [] as string[],
    ipAddresses: [] as string[],
  },
})

const signCertFiles = ref<UploadUserFile[]>([])
const signKeyFiles = ref<UploadUserFile[]>([])
const encCertFiles = ref<UploadUserFile[]>([])
const encKeyFiles = ref<UploadUserFile[]>([])

const updateSignCertFiles = ref<UploadUserFile[]>([])
const updateEncCertFiles = ref<UploadUserFile[]>([])
const selectedKeyStore = ref<any>(null)

const exportCSRForm = ref({
  keyType: 'sign' as 'sign' | 'enc',
  csrParams: {
    commonName: '',
    country: '',
    stateOrProvince: '',
    locality: '',
    org: '',
    orgUnit: '',
    emailAddress: '',
    dnsNames: [] as string[],
    ipAddresses: [] as string[],
  },
})

const hasTLCP = computed(() => keystores.value.some((k) => k.type === CertType.TLCP))

onMounted(() => fetchKeyStores())

async function fetchKeyStores() {
  loading.value = true
  try {
    const result = await keyStoreApi.list()
    keystores.value = result.keystores || []
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

function showExportCSRDialog(row: any) {
  selectedKeyStore.value = row
  exportCSRForm.value = {
    keyType: 'sign',
    csrParams: {
      commonName: '',
      country: '',
      stateOrProvince: '',
      locality: '',
      org: '',
      orgUnit: '',
      emailAddress: '',
      dnsNames: [],
      ipAddresses: [],
    },
  }
  showExportCSRDialog.value = true
}

async function exportCSR() {
  if (!exportCSRForm.value.csrParams.commonName) {
    ElMessage.error('请填写通用名称 (CN)')
    return
  }

  exportCSRLoading.value = true
  try {
    await keyStoreApi.exportCSR(selectedKeyStore.value.name, exportCSRForm.value)
    ElMessage.success('CSR导出成功')
    showExportCSRDialog.value = false
  } catch (err: any) {
    ElMessage.error(err.message || '导出失败')
  } finally {
    exportCSRLoading.value = false
  }
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

async function generateKeyStore() {
  if (!generateForm.value.name) {
    ElMessage.error('请填写密钥名称')
    return
  }
  if (!generateForm.value.certConfig.commonName) {
    ElMessage.error('请填写通用名称 (CN)')
    return
  }

  generateLoading.value = true
  try {
    await keyStoreApi.generate(generateForm.value)
    ElMessage.success('密钥生成成功')
    showGenerateDialog.value = false
    resetGenerateForm()
    fetchKeyStores()
  } catch (err: any) {
    ElMessage.error(err.message || '生成失败')
  } finally {
    generateLoading.value = false
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

function resetGenerateForm() {
  generateForm.value = {
    name: '',
    type: CertType.TLCP,
    protected: false,
    certConfig: {
      commonName: '',
      country: '',
      stateOrProvince: '',
      locality: '',
      org: '',
      orgUnit: '',
      emailAddress: '',
      years: 1,
      days: 0,
      keyAlgorithm: 'ecdsa',
      keyBits: 2048,
      dnsNames: [],
      ipAddresses: [],
    },
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
