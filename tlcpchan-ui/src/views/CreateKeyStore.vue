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
            <el-radio :value="CertType.IBC">标识密码 (IBC)</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="createForm.type !== CertType.IBC">
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
        </template>

        <!-- IBC(SM9) 身份材料：标识 + KGC 公共参数 + 三把用户私钥 -->
        <template v-if="createForm.type === CertType.IBC">
          <el-divider content-position="left">IBC 身份材料 (SM9，不使用 X.509 证书)</el-divider>

          <el-form-item label="标识文件" required>
            <el-upload
              v-model:file-list="identityFiles"
              :limit="1"
              :auto-upload="false"
              accept=".txt,.der,.pem"
            >
              <el-button type="primary">选择文件</el-button>
              <template #tip>
                <div class="el-upload__tip">本端标识文件（裸字节串文本或 Identifier DER），如 server@tlcpchan.local</div>
              </template>
            </el-upload>
          </el-form-item>

          <el-form-item label="KGC 公共参数">
            <el-upload
              v-model:file-list="paramsFiles"
              :limit="1"
              :auto-upload="false"
              accept=".pem,.der,.ibcparams"
            >
              <el-button type="primary">选择文件</el-button>
              <template #tip>
                <div class="el-upload__tip">本端 KGC 公共参数（PEM/DER/HEX/Base64）；服务端与双向认证客户端必需，仅服务端认证的客户端可省略</div>
              </template>
            </el-upload>
          </el-form-item>

          <el-form-item label="签名私钥">
            <el-upload
              v-model:file-list="ibcSignKeyFiles"
              :limit="1"
              :auto-upload="false"
              accept=".key,.pem"
            >
              <el-button type="primary">选择文件</el-button>
              <template #tip>
                <div class="el-upload__tip">hid=0x01，用于签名/验签（ServerKeyExchange、CertificateVerify）；服务端与双向认证客户端必需</div>
              </template>
            </el-upload>
          </el-form-item>

          <el-form-item label="加密私钥">
            <el-upload
              v-model:file-list="ibcEncKeyFiles"
              :limit="1"
              :auto-upload="false"
              accept=".key,.pem"
            >
              <el-button type="primary">选择文件</el-button>
              <template #tip>
                <div class="el-upload__tip">hid=0x03，用于解密预主密钥；使用 IBC 套件的服务端必需</div>
              </template>
            </el-upload>
          </el-form-item>

          <el-form-item label="密钥交换私钥">
            <el-upload
              v-model:file-list="ibcKexKeyFiles"
              :limit="1"
              :auto-upload="false"
              accept=".key,.pem"
            >
              <el-button type="primary">选择文件</el-button>
              <template #tip>
                <div class="el-upload__tip">hid=0x02，用于 SM9 密钥交换；使用 IBSDH 套件必需（注意勿与加密私钥混用）</div>
              </template>
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
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type UploadUserFile } from 'element-plus'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'
import { keystoreListLocation } from '@/constants/keystoreTab'

const route = useRoute()
const router = useRouter()
const createLoading = ref(false)

const createForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls' | 'ibc',
})

onMounted(() => {
  // 由密钥管理页面跳转而来时通过 query.type 预设密钥类型（白名单校验，非法或缺失时保持默认 TLCP）
  const presetType = route.query.type
  if (presetType === CertType.TLCP || presetType === CertType.TLS || presetType === CertType.IBC) {
    createForm.value.type = presetType
  }
})

const signCertFiles = ref<UploadUserFile[]>([])
const signKeyFiles = ref<UploadUserFile[]>([])
const encCertFiles = ref<UploadUserFile[]>([])
const encKeyFiles = ref<UploadUserFile[]>([])

// IBC 材料文件
const identityFiles = ref<UploadUserFile[]>([])
const paramsFiles = ref<UploadUserFile[]>([])
const ibcSignKeyFiles = ref<UploadUserFile[]>([])
const ibcEncKeyFiles = ref<UploadUserFile[]>([])
const ibcKexKeyFiles = ref<UploadUserFile[]>([])

/**
 * 返回密钥列表页，并停留在当前密钥类型所属的分类 Tab
 */
function goBack() {
  router.push(keystoreListLocation(createForm.value.type))
}

/**
 * 创建密钥存储
 * IBC 类型走 ibc-file 加载器，multipart 字段与后端契约一致；
 * 其余类型沿用证书密钥的创建流程。
 */
async function createKeyStore() {
  if (!createForm.value.name) {
    ElMessage.error('请填写密钥名称')
    return
  }

  // IBC 类型：标识文件必需，KGC 公共参数与三把私钥按用途选填
  if (createForm.value.type === CertType.IBC) {
    if (identityFiles.value.length === 0) {
      ElMessage.error('请上传 IBC 标识文件')
      return
    }

    createLoading.value = true
    try {
      const data: any = {
        name: createForm.value.name,
        type: CertType.IBC,
        loaderType: 'ibc-file',
        protected: false,
        identity: identityFiles.value[0]?.raw as File,
      }
      if (paramsFiles.value.length > 0) data.params = paramsFiles.value[0]?.raw as File
      if (ibcSignKeyFiles.value.length > 0) data.signKey = ibcSignKeyFiles.value[0]?.raw as File
      if (ibcEncKeyFiles.value.length > 0) data.encKey = ibcEncKeyFiles.value[0]?.raw as File
      if (ibcKexKeyFiles.value.length > 0) data.kexKey = ibcKexKeyFiles.value[0]?.raw as File

      await keyStoreApi.create(data)
      ElMessage.success('IBC 身份密钥创建成功')
      goBack()
    } catch (err: any) {
      ElMessage.error(err.message || '创建失败')
    } finally {
      createLoading.value = false
    }
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
      // 证书类型固定使用 file 加载器（与 CLI / 后端契约一致）
      loaderType: 'file',
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
