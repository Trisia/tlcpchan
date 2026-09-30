<template>
  <div class="keystores">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>密钥管理</span>
        </div>
      </template>

      <!-- 分类 Tab：PKI 证书类密钥与 IBC(SM9) 标识身份密钥，切换同步到 URL query -->
      <el-tabs v-model="activeTab">
        <el-tab-pane label="PKI" :name="KeystoreTab.PKI">
          <div class="tab-toolbar">
            <el-button type="success" @click="goToGenerate(CertType.TLCP)">
              <el-icon>
                <MagicStick />
              </el-icon>
              生成密钥
            </el-button>
            <el-button type="primary" @click="goToCreate(CertType.TLCP)">
              <el-icon>
                <Plus />
              </el-icon>
              导入密钥
            </el-button>
          </div>

          <el-table :data="certKeyStores" v-loading="loading" empty-text="暂无证书类密钥，可通过「生成密钥」或「导入密钥」创建">
            <el-table-column prop="name" label="名称" min-width="140" />
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
            <el-table-column label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" size="small" link @click="goToDetail(row)">详情</el-button>
                <el-button type="success" size="small" link @click="goToExportCSR(row)">导出 CSR</el-button>
                <el-button type="primary" size="small" link @click="goToUpdateCertificate(row)">更新证书</el-button>
                <el-button v-if="!row.protected" type="danger" size="small" link @click="remove(row.name)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="IBC" :name="KeystoreTab.IBC">
          <div class="tab-toolbar">
            <el-button type="success" @click="goToGenerate(CertType.IBC)">
              <el-icon>
                <MagicStick />
              </el-icon>
              生成 IBC 身份
            </el-button>
            <el-button type="primary" @click="goToCreate(CertType.IBC)">
              <el-icon>
                <Plus />
              </el-icon>
              导入 IBC 身份
            </el-button>
          </div>

          <el-table :data="ibcKeyStores" v-loading="loading"
            empty-text="暂无 IBC 身份密钥，可通过「生成 IBC 身份」或「导入 IBC 身份」创建">
            <el-table-column prop="name" label="名称" min-width="140" />
            <el-table-column label="标识" min-width="180">
              <template #default="{ row }">{{ row.ibc?.identity || '-' }}</template>
            </el-table-column>
            <el-table-column label="KGC 区域" width="150">
              <template #default="{ row }">{{ row.ibc?.districtName || '-' }}</template>
            </el-table-column>
            <el-table-column label="KGC 序号" width="100" align="center">
              <!-- 未提供 KGC 公共参数时后端返回 0，此时按“无”展示而非显示 0 -->
              <template #default="{ row }">{{ row.ibc?.districtName ? row.ibc?.districtSerial : '-' }}</template>
            </el-table-column>
            <el-table-column label="公共参数" width="110" align="center">
              <template #default="{ row }">
                <el-tag :type="row.ibc?.hasParams ? 'success' : 'info'" size="small">
                  {{ row.ibc?.hasParams ? '已提供' : '未提供' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="公共参数有效期" width="330">
              <template #default="{ row }">
                <span v-if="row.ibc?.hasParams">
                  {{ formatDateTime(row.ibc?.notBefore, '不限') }} ~ {{ formatDateTime(row.ibc?.notAfter, '不限') }}
                </span>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="用户私钥" min-width="230">
              <template #default="{ row }">
                <el-tag :type="row.ibc?.hasSignKey ? 'success' : 'info'" size="small" title="hid=0x01 签名用户私钥"
                  style="margin-right: 4px;">签名</el-tag>
                <el-tag :type="row.ibc?.hasEncryptKey ? 'success' : 'info'" size="small" title="hid=0x03 加密用户私钥"
                  style="margin-right: 4px;">加密</el-tag>
                <el-tag :type="row.ibc?.hasKeyExchangeKey ? 'success' : 'info'" size="small"
                  title="hid=0x02 密钥交换用户私钥">密钥交换</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" size="small" link @click="goToDetail(row)">详情</el-button>
                <el-button type="primary" size="small" link @click="goToUpdateCertificate(row)">替换材料</el-button>
                <el-button v-if="!row.protected" type="danger" size="small" link @click="remove(row.name)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, MagicStick } from '@element-plus/icons-vue'
import { keyStoreApi } from '@/api'
import { CertType } from '@/types'
import type { KeyStoreInfo } from '@/types'
import { KeystoreTab, filterKeyStoresByTab, normalizeKeyStoreTab } from '@/constants/keystoreTab'
import { formatDateTime } from '@/constants/datetime'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const keystores = ref<KeyStoreInfo[]>([])

// 当前分类 Tab，初值取自 URL query（非法或缺省时回退到 PKI）
const activeTab = ref<string>(normalizeKeyStoreTab(route.query.tab))

// PKI 分类：TLCP 的 SM2 双证书与 TLS 的 RSA/ECC 证书
const certKeyStores = computed(() => filterKeyStoresByTab(keystores.value, KeystoreTab.PKI))
// IBC 分类：SM9 标识身份密钥
const ibcKeyStores = computed(() => filterKeyStoresByTab(keystores.value, KeystoreTab.IBC))
// 证书列表中是否存在 TLCP 密钥，决定是否展示加密证书/密钥列
const hasTLCP = computed(() => keystores.value.some((k) => k.type === CertType.TLCP))

onMounted(() => fetchKeyStores())

// 浏览器前进/后退或外部跳转改变 query 时，把分类同步回组件
watch(
  () => route.query.tab,
  (raw) => {
    const tab = normalizeKeyStoreTab(raw)
    if (tab !== activeTab.value) activeTab.value = tab
  }
)

// 切换分类时写入 URL query，刷新或分享链接后仍停留在同一分类
watch(activeTab, (tab) => {
  if (normalizeKeyStoreTab(route.query.tab) === tab) return
  router.replace({ query: { ...route.query, tab } })
})

/**
 * 跳转到生成密钥页面
 * @param type 预设的密钥类型（tlcp / tls / ibc），由目标页面预选对应类型
 */
function goToGenerate(type: string) {
  router.push({ path: '/keystores/generate', query: { type } })
}

/**
 * 跳转到导入密钥页面
 * @param type 预设的密钥类型（tlcp / tls / ibc），由目标页面预选对应类型
 */
function goToCreate(type: string) {
  router.push({ path: '/keystores/create', query: { type } })
}

/**
 * 跳转到更新证书/替换材料页面
 * @param row 密钥存储信息，query 携带类型供目标页面区分证书与 IBC 表单
 */
function goToUpdateCertificate(row: KeyStoreInfo) {
  router.push({
    name: 'keystores-update',
    params: { name: row.name },
    query: { type: row.type }
  })
}

/**
 * 跳转到详情页面
 * @param row 密钥存储信息
 */
function goToDetail(row: KeyStoreInfo) {
  router.push({
    name: 'keystore-detail',
    params: { name: row.name }
  })
}

/**
 * 跳转到导出CSR页面
 * @param row 密钥存储信息（仅证书类密钥支持导出 CSR）
 */
function goToExportCSR(row: KeyStoreInfo) {
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

/* 每个分类 Tab 内的操作按钮区，与表格保持固定间距 */
.tab-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-bottom: 12px;
}
</style>