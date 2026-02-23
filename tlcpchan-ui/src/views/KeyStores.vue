<template>
  <div class="keystores">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>密钥管理</span>
          <div>
            <el-button type="success" @click="goToGenerate">
              <el-icon>
                <MagicStick />
              </el-icon>
              生成密钥
            </el-button>
            <el-button type="primary" @click="goToCreate">
              <el-icon>
                <Plus />
              </el-icon>
              创建密钥
            </el-button>
          </div>
        </div>
      </template>

      <el-table :data="keystores" v-loading="loading">
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="type" label="类型" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.type === CertType.TLCP ? 'primary' : 'success'">{{ row.type.toUpperCase()
            }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="签名证书/密钥" width="180">
          <template #default="{ row }">
            <span style="display: inline-block;margin-right: 5px;">
              <el-tag v-if="row.params['sign-cert']" size="small" type="success">有证书</el-tag>
              <el-tag v-else size="small" type="info">无证书</el-tag>
            </span>
            <span>
              <el-tag v-if="row.params['sign-key']" size="small" type="success">有密钥</el-tag>
              <el-tag v-else size="small" type="info">无密钥</el-tag>
            </span>
          </template>
        </el-table-column>
        <el-table-column v-if="hasTLCP" label="加密证书/密钥" width="180">
          <template #default="{ row }">
            <span v-if="row.type === CertType.TLCP">
              <span style="display: inline-block;margin-right: 5px;">
                <el-tag v-if="row.params['enc-cert']" size="small" type="success">有证书</el-tag>
                <el-tag v-else size="small" type="info">无证书</el-tag>
              </span>
              <span>
                <el-tag v-if="row.params['enc-key']" size="small" type="success">有密钥</el-tag>
                <el-tag v-else size="small" type="info">无密钥</el-tag>
              </span>
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>
               <el-table-column label="算法" width="100">
          <template #default="{ row }">
            {{ row.type === CertType.TLCP ? 'SM2' : 'ECDSA' }}
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="保护状态" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.protected" size="small" type="warning">受保护</el-tag>
            <el-tag v-else size="small">可删除</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="success" size="small" link @click="goToExportCSR(row)">导出 CSR</el-button>
            <el-button type="primary" size="small" link @click="goToUpdateCertificate(row)">更新证书</el-button>
            <el-button 
              v-if="!row.protected" 
              type="danger" 
              size="small" 
              link 
              @click="remove(row.name)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, MagicStick } from '@element-plus/icons-vue'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'

const router = useRouter()
const loading = ref(false)
const keystores = ref<any[]>([])

const hasTLCP = computed(() => keystores.value.some((k) => k.type === CertType.TLCP))

onMounted(() => fetchKeyStores())

/**
 * 跳转到创建密钥页面
 */
function goToCreate() {
  router.push('/keystores/create')
}

/**
 * 跳转到生成密钥页面
 */
function goToGenerate() {
  router.push('/keystores/generate')
}

/**
 * 跳转到更新证书页面
 * @param row 密钥存储信息
 */
function goToUpdateCertificate(row: any) {
  router.push({
    name: 'keystores-update',
    params: { name: row.name },
    query: { type: row.type }
  })
}

/**
 * 跳转到导出CSR页面
 * @param row 密钥存储信息
 */
function goToExportCSR(row: any) {
  router.push({
    name: 'keystores-export-csr',
    params: { name: row.name },
    query: { type: row.type }
  })
}

/**
 * 获取密钥列表
 */
async function fetchKeyStores() {
  loading.value = true
  try {
    const result = await keyStoreApi.list()
    keystores.value = result.keystores || []
  } catch (err) {
    console.error('获取密钥列表失败', err)
  } finally {
    loading.value = false
  }
}

/**
 * 格式化日期
 * @param dateStr 日期字符串
 */
function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

/**
 * 删除密钥存储
 * @param name 密钥存储名称
 */
function remove(name: string) {
  ElMessageBox.confirm('确定要删除此密钥吗？', '确认删除', { type: 'warning' })
    .then(async () => {
      try {
        await keyStoreApi.delete(name)
        ElMessage.success('密钥已删除')
        fetchKeyStores()
      } catch (err) {
        console.error('删除失败', err)
      }
    })
    .catch(() => { })
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
