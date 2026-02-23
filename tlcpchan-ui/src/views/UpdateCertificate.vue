<template>
  <div class="update-cert">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ path: '/keystores' }">密钥管理</el-breadcrumb-item>
      <el-breadcrumb-item>更新证书</el-breadcrumb-item>
    </el-breadcrumb>
    
    <el-card class="info-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>密钥库信息</span>
        </div>
      </template>
      <div class="info-content">
        <div class="info-item">
          <span class="info-label">密钥库名称：</span>
          <span class="info-value">{{ keyStoreName }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">类型：</span>
          <span class="info-value">{{ keyStoreType ? keyStoreType.toUpperCase() : '-' }}</span>
        </div>
      </div>
    </el-card>
    
    <el-card class="upload-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>签名证书</span>
        </div>
      </template>
      <el-upload 
        v-model:file-list="signCertFiles" 
        :limit="1" 
        :auto-upload="false" 
        accept=".crt,.pem"
        drag
      >
        <div class="upload-content">
          <el-icon class="upload-icon"><UploadFilled /></el-icon>
          <div class="upload-text">
            <p>点击或拖拽文件到此处上传</p>
            <p class="upload-tip">支持 .crt 或 .pem 格式</p>
          </div>
        </div>
      </el-upload>
    </el-card>
    
    <el-card v-if="keyStoreType === 'tlcp'" class="upload-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>加密证书</span>
        </div>
      </template>
      <el-upload 
        v-model:file-list="encCertFiles" 
        :limit="1" 
        :auto-upload="false" 
        accept=".crt,.pem"
        drag
      >
        <div class="upload-content">
          <el-icon class="upload-icon"><UploadFilled /></el-icon>
          <div class="upload-text">
            <p>点击或拖拽文件到此处上传</p>
            <p class="upload-tip">支持 .crt 或 .pem 格式</p>
          </div>
        </div>
      </el-upload>
    </el-card>
    
    <el-card class="action-card" style="margin-top: 16px;">
      <div class="action-content">
        <el-button @click="goBack">取消</el-button>
        <el-button type="primary" :loading="updateLoading" @click="updateCertificates">更新</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type UploadUserFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { keyStoreApi } from '@/api'

const router = useRouter()
const route = useRoute()
const updateLoading = ref(false)

const keyStoreName = computed(() => route.params.name as string)
const keyStoreType = computed(() => route.query.type as string)

const signCertFiles = ref<UploadUserFile[]>([])
const encCertFiles = ref<UploadUserFile[]>([])

/**
 * 返回密钥列表页
 */
function goBack() {
  router.push('/keystores')
}

/**
 * 更新证书
 */
async function updateCertificates() {
  if (signCertFiles.value.length === 0 && encCertFiles.value.length === 0) {
    ElMessage.error('请至少选择一个证书文件')
    return
  }

  updateLoading.value = true
  try {
    const data: any = {}
    if (signCertFiles.value.length > 0) {
      data.signCert = signCertFiles.value[0]?.raw as File
    }
    if (encCertFiles.value.length > 0) {
      data.encCert = encCertFiles.value[0]?.raw as File
    }

    await keyStoreApi.updateCertificates(keyStoreName.value, data)
    ElMessage.success('证书更新成功')
    goBack()
  } catch (err: any) {
    ElMessage.error(err.message || '更新失败')
  } finally {
    updateLoading.value = false
  }
}
</script>

<style scoped>
.update-cert {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
}

.info-card {
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
}

.info-content {
  display: flex;
  gap: 40px;
  padding: 8px 0;
}

.info-item {
  display: flex;
  align-items: center;
  font-size: 14px;
}

.info-label {
  color: #606266;
  font-weight: 500;
}

.info-value {
  color: #303133;
  font-weight: 600;
  margin-left: 8px;
}

.upload-card {
  transition: all 0.3s ease;
}

.upload-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.upload-content {
  text-align: center;
  padding: 30px 0;
}

.upload-icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 12px;
}

.upload-text p {
  margin: 8px 0;
  font-size: 14px;
  color: #606266;
}

.upload-tip {
  font-size: 12px;
  color: #909399 !important;
}

:deep(.el-upload-dragger) {
  width: 100%;
  border-radius: 8px;
}

.action-card {
  background: #f5f7fa;
}

.action-content {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 8px 0;
}
</style>
