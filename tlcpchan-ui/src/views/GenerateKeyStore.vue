<template>
  <div class="generate-keystore">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ path: '/keystores' }">密钥管理</el-breadcrumb-item>
      <el-breadcrumb-item>生成密钥</el-breadcrumb-item>
    </el-breadcrumb>
    
    <el-card class="form-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>生成密钥</span>
        </div>
      </template>
      
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
              <el-input-number v-model="generateForm.certConfig.years" :min="1" :max="100" style="width: 100%;" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="或 有效期(天)">
              <el-input-number v-model="generateForm.certConfig.days" :min="1" :max="36500" style="width: 100%;" />
            </el-form-item>
          </el-col>
        </el-row>

        <template v-if="generateForm.type === CertType.TLS">
          <el-divider content-position="left">密钥选项 (仅 TLS)</el-divider>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="密钥算法">
                <el-select v-model="generateForm.certConfig.keyAlgorithm" style="width: 100%;">
                  <el-option label="ECDSA" value="ecdsa" />
                  <el-option label="RSA" value="rsa" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="密钥位数 (仅 RSA)">
                <el-select v-model="generateForm.certConfig.keyBits" style="width: 100%;">
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
            style="width: 100%;" 
          />
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-select 
            v-model="generateForm.certConfig.ipAddresses" 
            multiple 
            filterable 
            allow-create
            placeholder="添加 IP 地址，如 192.168.1.1" 
            style="width: 100%;" 
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="goBack">取消</el-button>
        <el-button type="primary" :loading="generateLoading" @click="generateKeyStore">生成</el-button>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'

const router = useRouter()
const generateLoading = ref(false)

const generateForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls',
  protected: false as boolean,
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

/**
 * 返回密钥列表页
 */
function goBack() {
  router.push('/keystores')
}

/**
 * 生成密钥存储
 */
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
    goBack()
  } catch (err: any) {
    ElMessage.error(err.message || '生成失败')
  } finally {
    generateLoading.value = false
  }
}
</script>

<style scoped>
.generate-keystore {
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
