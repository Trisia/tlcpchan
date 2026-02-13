<template>
  <div class="certificates">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>证书管理</span>
          <div>
            <el-button type="success" @click="reload">
              <el-icon><Refresh /></el-icon>
              热更新证书
            </el-button>
            <el-button type="primary" @click="showCreateDialog = true">
              <el-icon><Plus /></el-icon>
              生成证书
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="certificates" v-loading="loading">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.type === 'tlcp' ? 'primary' : 'success'">{{ row.type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="subject" label="主题" />
        <el-table-column prop="issuer" label="颁发者" />
        <el-table-column prop="not_after" label="过期时间" width="180">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isExpiringSoon(row.not_after) }">{{ formatDate(row.not_after) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="public_key_algorithm" label="算法" width="100" />
        <el-table-column prop="is_ca" label="CA" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.is_ca" size="small" type="warning">CA</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="viewDetail(row)">详情</el-button>
            <el-button type="danger" size="small" link @click="remove(row.name)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="showCreateDialog" title="生成证书" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="证书类型">
          <el-select v-model="form.type">
            <el-option label="TLCP" value="tlcp" />
            <el-option label="TLS" value="tls" />
          </el-select>
        </el-form-item>
        <el-form-item label="证书名称" required>
          <el-input v-model="form.name" placeholder="server-sm2" />
        </el-form-item>
        <el-form-item label="通用名称" required>
          <el-input v-model="form.common_name" placeholder="localhost" />
        </el-form-item>
        <el-form-item label="DNS名称">
          <el-input v-model="dnsNamesStr" placeholder="localhost, *.example.com" />
        </el-form-item>
        <el-form-item label="IP地址">
          <el-input v-model="ipAddressesStr" placeholder="127.0.0.1, 192.168.1.1" />
        </el-form-item>
        <el-form-item label="有效期(天)">
          <el-input-number v-model="form.days" :min="1" :max="3650" />
        </el-form-item>
        <el-form-item label="CA证书">
          <el-select v-model="form.ca_name" clearable placeholder="留空则自签名">
            <el-option v-for="c in caCertificates" :key="c.name" :label="c.name" :value="c.name" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="generate">生成</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showDetailDialog" title="证书详情" width="500px">
      <el-descriptions v-if="selectedCert" :column="1" border>
        <el-descriptions-item label="名称">{{ selectedCert.name }}</el-descriptions-item>
        <el-descriptions-item label="类型">{{ selectedCert.type.toUpperCase() }}</el-descriptions-item>
        <el-descriptions-item label="主题">{{ selectedCert.subject }}</el-descriptions-item>
        <el-descriptions-item label="颁发者">{{ selectedCert.issuer }}</el-descriptions-item>
        <el-descriptions-item label="序列号">{{ selectedCert.serial_number }}</el-descriptions-item>
        <el-descriptions-item label="生效时间">{{ formatDate(selectedCert.not_before) }}</el-descriptions-item>
        <el-descriptions-item label="过期时间">{{ formatDate(selectedCert.not_after) }}</el-descriptions-item>
        <el-descriptions-item label="公钥算法">{{ selectedCert.public_key_algorithm }}</el-descriptions-item>
        <el-descriptions-item label="签名算法">{{ selectedCert.signature_algorithm }}</el-descriptions-item>
        <el-descriptions-item label="DNS名称">{{ selectedCert.dns_names?.join(', ') || '-' }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ selectedCert.ip_addresses?.join(', ') || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useCertificateStore } from '@/stores/certificate'
import { certificateApi } from '@/api'
import type { Certificate } from '@/types'

const store = useCertificateStore()

const loading = computed(() => store.loading)
const certificates = computed(() => store.certificates)
const caCertificates = computed(() => certificates.value.filter((c) => c.is_ca))

const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const selectedCert = ref<Certificate | null>(null)

const form = ref({
  type: 'tlcp' as 'tlcp' | 'tls',
  name: '',
  common_name: '',
  days: 365,
  ca_name: '',
})
const dnsNamesStr = ref('')
const ipAddressesStr = ref('')

onMounted(() => store.fetchCertificates())

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

function isExpiringSoon(dateStr: string): boolean {
  const date = new Date(dateStr)
  const now = new Date()
  const days = (date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  return days < 30
}

async function reload() {
  await store.reloadCertificates()
  ElMessage.success('证书已热更新')
}

function viewDetail(cert: Certificate) {
  selectedCert.value = cert
  showDetailDialog.value = true
}

async function remove(name: string) {
  await ElMessageBox.confirm('确定要删除此证书吗？', '确认删除', { type: 'warning' })
  await certificateApi.delete(name)
  await store.fetchCertificates()
  ElMessage.success('证书已删除')
}

async function generate() {
  if (!form.value.name || !form.value.common_name) {
    ElMessage.error('请填写必填项')
    return
  }
  await certificateApi.generate({
    ...form.value,
    dns_names: dnsNamesStr.value ? dnsNamesStr.value.split(',').map((s) => s.trim()) : undefined,
    ip_addresses: ipAddressesStr.value ? ipAddressesStr.value.split(',').map((s) => s.trim()) : undefined,
  })
  showCreateDialog.value = false
  await store.fetchCertificates()
  ElMessage.success('证书生成成功')
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.text-danger {
  color: #f56c6c;
}
</style>
