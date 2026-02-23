<template>
  <div class="create-keystore">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ path: '/keystores' }">密钥管理</el-breadcrumb-item>
      <el-breadcrumb-item>创建密钥</el-breadcrumb-item>
    </el-breadcrumb>
    
    <el-card class="form-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>创建密钥</span>
        </div>
      </template>
      
      <el-form :model="createForm" label-width="140px" style="max-width: 700px;">
        <el-form-item label="密钥名称" required>
          <el-input v-model="createForm.name" placeholder="my-server-key" style="width: 100%;" />
        </el-form-item>
        
        <el-form-item label="类型" required>
          <el-radio-group v-model="createForm.type">
            <el-radio :value="CertType.TLCP">国密 (TLCP)</el-radio>
            <el-radio :value="CertType.TLS">国际 (TLS)</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-divider content-position="left">签名证书和密钥</el-divider>
        
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
        
        <template v-if="createForm.type === CertType.TLCP">
          <el-divider content-position="left">加密证书和密钥 (仅 TLCP)</el-divider>
          
          <el-form-item label="加密证书" required>
            <el-upload 
              v-model:file-list="encCertFiles" 
              :limit="1" 
              :auto-upload="false" 
              accept=".crt,.pem"
            >
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
          
          <el-form-item label="加密密钥" required>
            <el-upload 
              v-model:file-list="encKeyFiles" 
              :limit="1" 
              :auto-upload="false" 
              accept=".key,.pem"
            >
              <el-button type="primary">选择文件</el-button>
            </el-upload>
          </el-form-item>
        </template>
      </el-form>
      
      <template #footer>
        <el-button @click="goBack">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="createKeyStore">创建</el-button>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type UploadUserFile } from 'element-plus'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'

const router = useRouter()
const createLoading = ref(false)

const createForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls',
})

const signCertFiles = ref<UploadUserFile[]>([])
const signKeyFiles = ref<UploadUserFile[]>([])
const encCertFiles = ref<UploadUserFile[]>([])
const encKeyFiles = ref<UploadUserFile[]>([])

/**
 * 返回密钥列表页
 */
function goBack() {
  router.push('/keystores')
}

/**
 * 创建密钥存储
 */
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
      signCert: signCertFiles.value[0]?.raw as File,
      signKey: signKeyFiles.value[0]?.raw as File,
    }
    if (createForm.value.type === CertType.TLCP) {
      data.encCert = encCertFiles.value[0]?.raw as File
      data.encKey = encKeyFiles.value[0]?.raw as File
    }

    await keyStoreApi.create(data)
    ElMessage.success('密钥创建成功')
    goBack()
  } catch (err: any) {
    ElMessage.error(err.message || '创建失败')
  } finally {
    createLoading.value = false
  }
}
</script>

<style scoped>
.create-keystore {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
}
</style>
