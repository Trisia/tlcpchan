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
      <el-alert
        title="提示：根 CA 证书用于签发服务器和客户端证书"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      />
      <el-form :model="generateForm" label-width="120px">
        <el-divider content-position="left">证书类型</el-divider>
        <el-form-item label="证书类型" required>
          <el-radio-group v-model="generateForm.type">
            <el-radio value="tlcp">国密 SM2 (TLCP)</el-radio>
            <el-radio value="tls">国际 RSA (TLS)</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-divider content-position="left">证书主体信息</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="国家代码 (C)">
              <el-input v-model="generateForm.country" placeholder="CN" maxlength="2" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="省份/直辖市 (ST)">
              <el-input v-model="generateForm.stateOrProvince" placeholder="北京市" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="城市/地区 (L)">
              <el-input v-model="generateForm.locality" placeholder="北京市" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="组织名称 (O)">
              <el-input v-model="generateForm.org" placeholder="公司名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="组织单位 (OU)">
              <el-input v-model="generateForm.orgUnit" placeholder="IT 部门" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="通用名称 (CN)" required>
              <el-input v-model="generateForm.commonName" placeholder="my-root-ca" />
              <template #label>
                <span>通用名称 (CN) <el-tooltip content="证书的唯一标识，建议使用根证书名称" placement="top"><el-icon><QuestionFilled /></el-icon></el-tooltip></span>
              </template>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱地址 (E)">
              <el-input v-model="generateForm.emailAddress" placeholder="admin@company.com" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider content-position="left">证书有效期</el-divider>
        <el-form-item label="有效期" required>
          <el-select v-model="generateForm.validityPeriod" placeholder="请选择有效期" style="width: 100%;">
            <el-option label="1 年 (365 天)" value="1y" />
            <el-option label="3 年 (1095 天)" value="3y" />
            <el-option label="5 年 (1825 天)" value="5y" />
            <el-option label="10 年 (3650 天)" value="10y" />
            <el-option label="20 年 (7300 天)" value="20y" />
            <el-option label="50 年 (18250 天)" value="50y" />
            <el-option label="100 年 (36500 天)" value="100y" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">其他配置</el-divider>
        <el-form-item label="密钥长度">
          <el-select v-model="generateForm.keySize" placeholder="请选择密钥长度" style="width: 100%;" v-if="generateForm.type === 'tls'">
            <el-option label="2048 位 (推荐)" :value="2048" />
            <el-option label="4096 位" :value="4096" />
          </el-select>
          <el-input v-else value="SM2 256 位 (国密标准)" disabled style="width: 100%;" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGenerateDialog = false">取消</el-button>
        <el-button type="primary" :loading="generateLoading" @click="generateRootCA">生成根证书</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { Plus, MagicStick, QuestionFilled } from '@element-plus/icons-vue'
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
  country: 'CN',
  stateOrProvince: '',
  locality: '',
  org: 'tlcpchan',
  orgUnit: '',
  emailAddress: '',
  validityPeriod: '10y',
  keySize: 256,
})

onMounted(() => fetchTrustedCerts())

watch(() => generateForm.value.type, (newType) => {
  generateForm.value.keySize = newType === 'tlcp' ? 256 : 2048
})

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
  if (!generateForm.value.validityPeriod) {
    ElMessage.error('请选择证书有效期')
    return
  }
  if (!generateForm.value.keySize) {
    ElMessage.error('请选择密钥长度')
    return
  }

  generateLoading.value = true
  try {
    const period = generateForm.value.validityPeriod
    const data = {
      ...generateForm.value,
      years: period.endsWith('y') ? parseInt(period) : 0,
      days: period.endsWith('y') ? 0 : parseInt(period),
    }
    
    await rootCertApi.generate(data)
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
    await ElMessageBox.prompt(
      `请输入证书文件名 <span style="color: red; font-weight: bold;">${name}</span> 确认删除`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        dangerouslyUseHTMLString: true,
        inputPattern: new RegExp(`^${name}$`),
        inputErrorMessage: '文件名不匹配，删除已取消',
      }
    )
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
    country: 'CN',
    stateOrProvince: '',
    locality: '',
    org: 'tlcpchan',
    orgUnit: '',
    emailAddress: '',
    validityPeriod: '10y',
    keySize: 256,
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
