<template>
  <div class="export-csr">
    <el-breadcrumb separator="/">
      <el-breadcrumb-item :to="{ path: '/keystores' }">密钥管理</el-breadcrumb-item>
      <el-breadcrumb-item>导出证书请求 (CSR)</el-breadcrumb-item>
    </el-breadcrumb>
    
    <el-card class="form-card" style="margin-top: 16px;">
      <template #header>
        <div class="card-header">
          <span>导出证书请求 (CSR) - {{ keyStoreName }}</span>
        </div>
      </template>
      
      <el-alert 
        v-if="keyStoreType" 
        :title="`类型: ${keyStoreType.toUpperCase()}`" 
        type="info" 
        :closable="false"
        style="margin-bottom: 16px;"
      />
      
      <el-form :model="exportCSRForm" label-width="140px">
        <el-form-item v-if="keyStoreType === 'tlcp'" label="密钥类型" required>
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
            style="width: 100%;" 
          />
        </el-form-item>
        <el-form-item label="IP 地址">
          <el-select 
            v-model="exportCSRForm.csrParams.ipAddresses" 
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
        <el-button type="primary" :loading="exportCSRLoading" @click="exportCSR">导出</el-button>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { keyStoreApi } from '@/api'

const router = useRouter()
const route = useRoute()
const exportCSRLoading = ref(false)

const keyStoreName = computed(() => route.params.name as string)
const keyStoreType = computed(() => route.query.type as string)

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

/**
 * 返回密钥列表页
 */
function goBack() {
  router.push('/keystores')
}

/**
 * 导出CSR
 */
async function exportCSR() {
  if (!exportCSRForm.value.csrParams.commonName) {
    ElMessage.error('请填写通用名称 (CN)')
    return
  }

  exportCSRLoading.value = true
  try {
    await keyStoreApi.exportCSR(keyStoreName.value, exportCSRForm.value)
    ElMessage.success('CSR导出成功')
    goBack()
  } catch (err: any) {
    ElMessage.error(err.message || '导出失败')
  } finally {
    exportCSRLoading.value = false
  }
}
</script>

<style scoped>
.export-csr {
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
