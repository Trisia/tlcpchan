<template>
  <div class="update-cert">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="keystoreListLocation(keyStoreType)">密钥管理</el-breadcrumb-item>
      <el-breadcrumb-item>{{ keyStoreType === 'ibc' ? '替换 IBC 材料' : '更新证书' }}</el-breadcrumb-item>
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
    
    <el-card v-if="keyStoreType !== 'ibc'" class="upload-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>签名证书和密钥</span>
        </div>
      </template>
      
      <div class="upload-section">
        <div class="upload-item">
          <label class="upload-label">签名证书</label>
          <el-upload 
            v-model:file-list="signCertFiles" 
            :limit="1" 
            :auto-upload="false" 
            accept=".crt,.pem"
            drag
            @change="handleSignCertChange"
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">支持 .crt 或 .pem 格式</p>
              </div>
            </div>
          </el-upload>
        </div>
        
        <div class="upload-item">
          <label class="upload-label">签名密钥（可选）</label>
          <el-upload 
            v-model:file-list="signKeyFiles" 
            :limit="1" 
            :auto-upload="false" 
            accept=".key,.pem"
            drag
            @change="handleSignKeyChange"
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">支持 .key 或 .pem 格式</p>
              </div>
            </div>
          </el-upload>
        </div>
      </div>
    </el-card>
    
    <el-card v-if="keyStoreType === 'tlcp'" class="upload-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>加密证书和密钥</span>
        </div>
      </template>
      
      <div class="upload-section">
        <div class="upload-item">
          <label class="upload-label">加密证书</label>
          <el-upload 
            v-model:file-list="encCertFiles" 
            :limit="1" 
            :auto-upload="false" 
            accept=".crt,.pem"
            drag
            @change="handleEncCertChange"
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">支持 .crt 或 .pem 格式</p>
              </div>
            </div>
          </el-upload>
        </div>
        
        <div class="upload-item">
          <label class="upload-label">加密密钥（可选）</label>
          <el-upload 
            v-model:file-list="encKeyFiles" 
            :limit="1" 
            :auto-upload="false" 
            accept=".key,.pem"
            drag
            @change="handleEncKeyChange"
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">支持 .key 或 .pem 格式</p>
              </div>
            </div>
          </el-upload>
        </div>
      </div>
    </el-card>
    
    <!-- IBC 身份材料替换（仅 ibc 类型） -->
    <el-card v-if="keyStoreType === 'ibc'" class="upload-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>IBC 身份材料（所有文件可选，仅上传需要替换的材料）</span>
        </div>
      </template>

      <div class="upload-section">
        <div class="upload-item">
          <label class="upload-label">标识文件</label>
          <el-upload
            v-model:file-list="identityFiles"
            :limit="1"
            :auto-upload="false"
            accept=".txt,.der,.pem"
            drag
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">本端标识，如 server@tlcpchan.local</p>
              </div>
            </div>
          </el-upload>
        </div>

        <div class="upload-item">
          <label class="upload-label">KGC 公共参数</label>
          <el-upload
            v-model:file-list="paramsFiles"
            :limit="1"
            :auto-upload="false"
            accept=".pem,.der,.ibcparams"
            drag
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">PEM/DER/HEX/Base64 格式</p>
              </div>
            </div>
          </el-upload>
        </div>
      </div>

      <div class="upload-section" style="margin-top: 24px;">
        <div class="upload-item">
          <label class="upload-label">签名私钥 (hid=0x01)</label>
          <el-upload
            v-model:file-list="ibcSignKeyFiles"
            :limit="1"
            :auto-upload="false"
            accept=".key,.pem"
            drag
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">用于签名/验签</p>
              </div>
            </div>
          </el-upload>
        </div>

        <div class="upload-item">
          <label class="upload-label">加密私钥 (hid=0x03)</label>
          <el-upload
            v-model:file-list="ibcEncKeyFiles"
            :limit="1"
            :auto-upload="false"
            accept=".key,.pem"
            drag
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">用于解密预主密钥</p>
              </div>
            </div>
          </el-upload>
        </div>

        <div class="upload-item">
          <label class="upload-label">密钥交换私钥 (hid=0x02)</label>
          <el-upload
            v-model:file-list="ibcKexKeyFiles"
            :limit="1"
            :auto-upload="false"
            accept=".key,.pem"
            drag
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><UploadFilled /></el-icon>
              <div class="upload-text">
                <p>点击或拖拽文件到此处上传</p>
                <p class="upload-tip">用于 SM9 密钥交换，勿与加密私钥混用</p>
              </div>
            </div>
          </el-upload>
        </div>
      </div>
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
import { keystoreListLocation } from '@/constants/keystoreTab'

const router = useRouter()
const route = useRoute()
const updateLoading = ref(false)

const keyStoreName = computed(() => route.params.name as string)
const keyStoreType = computed(() => route.query.type as string)

const signCertFiles = ref<UploadUserFile[]>([])
const signKeyFiles = ref<UploadUserFile[]>([])
const encCertFiles = ref<UploadUserFile[]>([])
const encKeyFiles = ref<UploadUserFile[]>([])

// IBC 材料替换文件
const identityFiles = ref<UploadUserFile[]>([])
const paramsFiles = ref<UploadUserFile[]>([])
const ibcSignKeyFiles = ref<UploadUserFile[]>([])
const ibcEncKeyFiles = ref<UploadUserFile[]>([])
const ibcKexKeyFiles = ref<UploadUserFile[]>([])

/**
 * 处理签名证书变更
 */
function handleSignCertChange() {
  // 如果签名证书被清除，也清除签名密钥
  if (signCertFiles.value.length === 0 && signKeyFiles.value.length > 0) {
    signKeyFiles.value = []
  }
}

/**
 * 处理签名密钥变更
 */
function handleSignKeyChange() {
  // 如果签名密钥被清除，保持签名证书
}

/**
 * 处理加密证书变更
 */
function handleEncCertChange() {
  if (encCertFiles.value.length === 0 && encKeyFiles.value.length > 0) {
    encKeyFiles.value = []
  }
}

/**
 * 处理加密密钥变更
 */
function handleEncKeyChange() {
  // 如果加密密钥被清除，保持加密证书
}

/**
 * 返回密钥列表页，并停留在当前密钥类型所属的分类 Tab
 */
function goBack() {
  router.push(keystoreListLocation(keyStoreType.value))
}

/**
 * 更新证书 / IBC 材料
 * ibc 类型上传 标识/参数/三把私钥（全部可选，至少一个）；
 * 证书类型沿用原有校验规则。
 */
async function updateCertificates() {
  // IBC 类型：五个材料字段全部可选，但至少要上传一个
  if (keyStoreType.value === 'ibc') {
    const data: any = {}
    if (identityFiles.value.length > 0) data.identity = identityFiles.value[0]?.raw as File
    if (paramsFiles.value.length > 0) data.params = paramsFiles.value[0]?.raw as File
    if (ibcSignKeyFiles.value.length > 0) data.signKey = ibcSignKeyFiles.value[0]?.raw as File
    if (ibcEncKeyFiles.value.length > 0) data.encKey = ibcEncKeyFiles.value[0]?.raw as File
    if (ibcKexKeyFiles.value.length > 0) data.kexKey = ibcKexKeyFiles.value[0]?.raw as File

    if (Object.keys(data).length === 0) {
      ElMessage.error('请至少选择一个要替换的材料文件')
      return
    }

    updateLoading.value = true
    try {
      await keyStoreApi.updateCertificates(keyStoreName.value, data)
      ElMessage.success('IBC 材料替换成功')
      goBack()
    } catch (err: any) {
      ElMessage.error(err.message || '更新失败')
    } finally {
      updateLoading.value = false
    }
    return
  }

  if (signCertFiles.value.length === 0 && encCertFiles.value.length === 0) {
    ElMessage.error('请至少选择一个证书文件')
    return
  }
  
  // 验证：如果上传了签名密钥，必须同时上传签名证书
  if (signKeyFiles.value.length > 0 && signCertFiles.value.length === 0) {
    ElMessage.error('上传签名密钥时必须同时上传签名证书')
    return
  }
  
  // 验证：如果上传了加密密钥，必须同时上传加密证书
  if (encKeyFiles.value.length > 0 && encCertFiles.value.length === 0) {
    ElMessage.error('上传加密密钥时必须同时上传加密证书')
    return
  }
  
  // 验证：TLCP 类型必须同时上传签名证书和加密证书
  if (keyStoreType.value === 'tlcp') {
    if (signCertFiles.value.length === 0 || encCertFiles.value.length === 0) {
      ElMessage.error('TLCP 类型必须同时上传签名证书和加密证书')
      return
    }
  }

  updateLoading.value = true
  try {
    const data: any = {}
    if (signCertFiles.value.length > 0) {
      data.signCert = signCertFiles.value[0]?.raw as File
    }
    if (signKeyFiles.value.length > 0) {
      data.signKey = signKeyFiles.value[0]?.raw as File
    }
    if (encCertFiles.value.length > 0) {
      data.encCert = encCertFiles.value[0]?.raw as File
    }
    if (encKeyFiles.value.length > 0) {
      data.encKey = encKeyFiles.value[0]?.raw as File
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

.upload-section {
  display: flex;
  gap: 24px;
}

.upload-item {
  flex: 1;
}

.upload-label {
  display: block;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
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
