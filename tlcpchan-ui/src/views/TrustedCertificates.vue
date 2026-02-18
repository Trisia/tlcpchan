<template>
  <div class="trusted">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>信任证书管理</span>
          <el-button type="primary" @click="showUploadDialog = true">
            <el-icon><Plus /></el-icon>
            上传证书
          </el-button>
        </div>
      </template>

      <el-table :data="trustedCerts" v-loading="loading">
        <el-table-column prop="name" label="文件名" />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="subject" label="主题" min-width="200" show-overflow-tooltip />
        <el-table-column prop="issuer" label="颁发者" min-width="200" show-overflow-tooltip />
        <el-table-column label="是否为CA" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.isCA" size="small" type="success">是</el-tag>
            <el-tag v-else size="small" type="info">否</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="expiresAt" label="过期时间" width="180">
          <template #default="{ row }">
            {{ row.expiresAt ? formatDate(row.expiresAt) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="danger" size="small" link @click="remove(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showUploadDialog" title="上传信任证书" width="500px">
      <el-form label-width="120px">
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
        <el-button type="primary" :loading="uploadLoading" @click="uploadCert">上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { trustedApi } from '@/api'

const loading = ref(false)
const trustedCerts = ref<any[]>([])

const showUploadDialog = ref(false)
const uploadLoading = ref(false)
const certFiles = ref<UploadUserFile[]>([])

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

async function uploadCert() {
  if (certFiles.value.length === 0) {
    ElMessage.error('请选择证书文件')
    return
  }

  uploadLoading.value = true
  try {
    await trustedApi.upload(certFiles.value[0].raw as File)
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

function remove(name: string) {
  ElMessageBox.confirm('确定要删除此信任证书吗？', '确认删除', { type: 'warning' })
    .then(async () => {
      try {
        await trustedApi.delete(name)
        ElMessage.success('信任证书已删除')
        fetchTrustedCerts()
      } catch (err) {
        console.error('删除失败', err)
      }
    })
    .catch(() => {})
}

function resetUploadForm() {
  certFiles.value = []
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
