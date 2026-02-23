<template>
  <div class="trusted">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>信任证书管理</span>
          <div>
            <el-button type="success" @click="showGenerateDialog = true">
              <el-icon><MagicStick /></el-icon>
              生成根证书
            </el-button>
            <el-button type="primary" @click="showUploadDialog = true">
              <el-icon><Plus /></el-icon>
              上传证书
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="trustedCerts" v-loading="loading">
        <el-table-column prop="filename" label="文件名" width="200" fixed="left" />
        <el-table-column prop="keyType" label="密钥类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.keyType }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="subject" label="主题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="issuer" label="颁发者" min-width="200" show-overflow-tooltip />
        <el-table-column label="有效期" width="320">
          <template #default="{ row }">
            {{ formatDate(row.notBefore) }} - {{ formatDate(row.notAfter) }}
          </template>
        </el-table-column>
        <el-table-column prop="serialNumber" label="序列号" width="150" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tooltip :content="row.serialNumber" placement="top">
              <span>{{ truncateSerialNumber(row.serialNumber) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="CA" width="60" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.isCA" size="small" type="success">是</el-tag>
            <el-tag v-else size="small" type="info">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="showDetail(row)">详情</el-button>
            <el-button type="success" size="small" link @click="download(row.filename)">下载</el-button>
            <el-button type="danger" size="small" link @click="remove(row.filename)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showDetailDialog" title="证书详情" width="700px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="文件名" label-align="right" label-class-name="detail-label">{{ currentCert?.filename }}</el-descriptions-item>
        <el-descriptions-item label="主题" label-align="right" label-class-name="detail-label">{{ currentCert?.subject }}</el-descriptions-item>
        <el-descriptions-item label="颁发者" label-align="right" label-class-name="detail-label">{{ currentCert?.issuer }}</el-descriptions-item>
        <el-descriptions-item label="生效时间" label-align="right" label-class-name="detail-label">{{ currentCert?.notBefore ? formatDate(currentCert.notBefore) : '' }}</el-descriptions-item>
        <el-descriptions-item label="过期时间" label-align="right" label-class-name="detail-label">{{ currentCert?.notAfter ? formatDate(currentCert.notAfter) : '' }}</el-descriptions-item>
        <el-descriptions-item label="密钥类型" label-align="right" label-class-name="detail-label">{{ currentCert?.keyType }}</el-descriptions-item>
        <el-descriptions-item label="序列号" label-align="right" label-class-name="detail-label">{{ currentCert?.serialNumber }}</el-descriptions-item>
        <el-descriptions-item label="版本" label-align="right" label-class-name="detail-label">v{{ currentCert?.version }}</el-descriptions-item>
        <el-descriptions-item label="CA 证书" label-align="right" label-class-name="detail-label">{{ currentCert?.isCA ? '是' : '否' }}</el-descriptions-item>
        <el-descriptions-item label="密钥用途" label-align="right" label-class-name="detail-label">
          <el-tag v-for="usage in currentCert?.keyUsage" :key="usage" size="small" style="margin: 2px">
            {{ usage }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showUploadDialog" title="上传信任证书" width="500px">
      <el-form label-width="120px">
        <el-form-item label="文件名" required>
          <el-input v-model="uploadForm.filename" placeholder="请输入证书文件名（如：my-root-ca.crt）" />
        </el-form-item>
        <el-form-item label="证书文件" required>
          <el-upload
            v-model:file-list="certFiles"
            :limit="1"
            :auto-upload="false"
            accept=".crt,.pem,.cer"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">支持 .crt、.pem、.cer 格式的证书文件</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="uploadCert">上传确认</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showGenerateDialog" title="生成根 CA 证书" width="700px">
      <el-form :model="generateForm" label-width="140px">
        <el-divider content-position="left">证书类型</el-divider>
        <el-form-item label="类型" required>
          <el-radio-group v-model="generateForm.type">
            <el-radio value="tlcp">TLCP (SM2)</el-radio>
            <el-radio value="tls">TLS (RSA)</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-divider content-position="left">证书主体 (DN)</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="国家 (C)">
              <el-input v-model="generateForm.country" placeholder="CN" maxlength="2" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="省/州 (ST)">
              <el-input v-model="generateForm.stateOrProvince" placeholder="Beijing" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="地区 (L)">
              <el-input v-model="generateForm.locality" placeholder="Haidian" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="组织 (O)">
              <el-input v-model="generateForm.org" placeholder="Example Org" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="组织单位 (OU)">
              <el-input v-model="generateForm.orgUnit" placeholder="IT" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="通用名称 (CN)" required>
              <el-input v-model="generateForm.commonName" placeholder="my-root-ca" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱地址">
              <el-input v-model="generateForm.emailAddress" placeholder="admin@example.com" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">有效期</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="有效期(年)">
              <el-input-number v-model="generateForm.years" :min="1" :max="100" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="或 有效期(天)">
              <el-input-number v-model="generateForm.days" :min="1" :max="36500" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="showGenerateDialog = false">取消</el-button>
        <el-button type="primary" :loading="generateLoading" @click="generateRootCA">生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { Plus, MagicStick } from '@element-plus/icons-vue'
import { trustedApi, rootCertApi } from '@/api'

const loading = ref(false)
const trustedCerts = ref<any[]>([])

const showUploadDialog = ref(false)
const showGenerateDialog = ref(false)
const showDetailDialog = ref(false)
const uploadLoading = ref(false)
const generateLoading = ref(false)
const currentCert = ref<any>(null)

const certFiles = ref<UploadUserFile[]>([])

const uploadForm = ref({
  filename: '',
})

const generateForm = ref({
  type: 'tlcp',
  commonName: 'tlcpchan-root-ca',
  country: '',
  stateOrProvince: '',
  locality: '',
  org: 'tlcpchan',
  orgUnit: '',
  emailAddress: '',
  years: 10,
  days: 0,
})

onMounted(() => fetchTrustedCerts())

async function fetchTrustedCerts() {
  loading.value = true
  try {
    trustedCerts.value = await trustedApi.list()
  } catch (err) {
    console.error('获取信任证书列表失败', err)
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

function truncateSerialNumber(serial: string): string {
  if (serial.length <= 20) {
    return serial
  }
  return serial.substring(0, 10) + '...' + serial.substring(serial.length - 10)
}

function showDetail(cert: any) {
  currentCert.value = cert
  showDetailDialog.value = true
}

async function download(filename: string) {
  try {
    await trustedApi.download(filename)
    ElMessage.success('证书下载成功')
  } catch (err: any) {
    ElMessage.error(err.message || '下载失败')
  }
}

async function uploadCert() {
  if (!uploadForm.value.filename) {
    ElMessage.error('请输入文件名')
    return
  }
  if (certFiles.value.length === 0) {
    ElMessage.error('请选择证书文件')
    return
  }

  uploadLoading.value = true
  try {
    await rootCertApi.add(uploadForm.value.filename, certFiles.value[0]!.raw as File)
    ElMessage.success('证书上传成功')
    showUploadDialog.value = false
    resetUploadForm()
    fetchTrustedCerts()
  } catch (err: any) {
    ElMessage.error(err.message || '上传失败')
  } finally {
    uploadLoading.value = false
  }
}

async function generateRootCA() {
  if (!generateForm.value.commonName) {
    ElMessage.error('请填写通用名称 (CN)')
    return
  }

  generateLoading.value = true
  try {
    await rootCertApi.generate(generateForm.value)
    ElMessage.success('根证书生成成功')
    showGenerateDialog.value = false
    resetGenerateForm()
    fetchTrustedCerts()
  } catch (err: any) {
    ElMessage.error(err.message || '生成失败')
  } finally {
    generateLoading.value = false
  }
}

async function remove(name: string) {
  try {
    await ElMessageBox.prompt('请输入证书文件名确认删除', '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: new RegExp(`^${name}$`),
      inputErrorMessage: '文件名不匹配，删除已取消',
    })
    await trustedApi.delete(name)
    ElMessage.success('信任证书已删除')
    fetchTrustedCerts()
  } catch (err) {
    if (err !== 'cancel') {
      console.error('删除失败', err)
    }
  }
}

function resetUploadForm() {
  uploadForm.value.filename = ''
  certFiles.value = []
}

function resetGenerateForm() {
  generateForm.value = {
    type: 'tlcp',
    commonName: 'tlcpchan-root-ca',
    country: '',
    stateOrProvince: '',
    locality: '',
    org: 'tlcpchan',
    orgUnit: '',
    emailAddress: '',
    years: 10,
    days: 0,
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  width: 120px;
}
</style>
