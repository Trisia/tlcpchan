<template>
  <div class="ibcparams">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>IBC 信任池</span>
          <div>
            <el-button type="warning" @click="reloadPool" :loading="reloadLoading">
              <el-icon><Refresh /></el-icon>
              重载信任池
            </el-button>
            <el-button type="success" @click="showGenerateDialog = true">
              <el-icon><MagicStick /></el-icon>
              生成测试 KGC
            </el-button>
            <el-button type="primary" @click="showUploadDialog = true">
              <el-icon><Plus /></el-icon>
              上传公共参数
            </el-button>
          </div>
        </div>
      </template>

      <el-alert
        title="IBC 信任池等价于根证书库：必须通过可信渠道获取 KGC 公共参数后再入库，绝不直接信任对端下发的参数。修改信任池后需重载相关实例才能生效。"
        type="warning"
        :closable="false"
        style="margin-bottom: 16px;"
      />

      <el-table :data="ibcParams" v-loading="loading">
        <el-table-column prop="filename" label="文件名" width="220" fixed="left" show-overflow-tooltip />
        <el-table-column prop="districtName" label="KGC 区域" min-width="160" show-overflow-tooltip />
        <el-table-column prop="districtSerial" label="序号" width="80" align="center" />
        <el-table-column label="生效时间" width="180">
          <template #default="{ row }">{{ row.notBefore ? formatDate(row.notBefore) : '不限' }}</template>
        </el-table-column>
        <el-table-column label="失效时间" width="180">
          <template #default="{ row }">{{ row.notAfter ? formatDate(row.notAfter) : '不限' }}</template>
        </el-table-column>
        <el-table-column label="签名主公钥指纹" min-width="150">
          <template #default="{ row }">
            <el-tooltip :content="row.signKeyFingerprint" placement="top">
              <span>{{ truncateFingerprint(row.signKeyFingerprint) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="加密主公钥指纹" min-width="150">
          <template #default="{ row }">
            <el-tooltip :content="row.encKeyFingerprint" placement="top">
              <span>{{ truncateFingerprint(row.encKeyFingerprint) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="success" size="small" link @click="download(row.filename)">下载</el-button>
            <el-button type="danger" size="small" link @click="remove(row.filename)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 上传 KGC 公共参数 -->
    <el-dialog v-model="showUploadDialog" title="上传 KGC 公共参数" width="500px">
      <el-form label-width="120px">
        <el-form-item label="文件名" required>
          <el-input v-model="uploadForm.filename" placeholder="请输入文件名（如：my-kgc.pem）" />
        </el-form-item>
        <el-form-item label="公共参数文件" required>
          <el-upload
            v-model:file-list="paramFiles"
            :limit="1"
            :auto-upload="false"
            accept=".pem,.der,.ibcparams"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">支持 .pem、.der、.ibcparams 格式的 KGC 公共参数文件</div>
            </template>
          </el-upload>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" :loading="uploadLoading" @click="uploadParam">上传确认</el-button>
      </template>
    </el-dialog>

    <!-- 生成测试 KGC 公共参数 -->
    <el-dialog v-model="showGenerateDialog" title="生成测试 KGC 公共参数" width="520px">
      <el-alert
        title="仅用于本地联调：生成的主密钥保存在服务端工作目录，不提供下载。生产环境请由外部 KGC 派生后导入。"
        type="warning"
        :closable="false"
        style="margin-bottom: 20px;"
      />
      <el-form :model="generateForm" label-width="140px">
        <el-form-item label="KGC 区域" required>
          <el-input v-model="generateForm.districtName" placeholder="tlcpchan.local" />
        </el-form-item>
        <el-form-item label="KGC 序号" required>
          <el-input-number v-model="generateForm.districtSerial" :min="1" :max="65535" style="width: 100%;" />
        </el-form-item>
        <el-form-item label="有效期(年)" required>
          <el-input-number v-model="generateForm.years" :min="1" :max="100" style="width: 100%;" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showGenerateDialog = false">取消</el-button>
        <el-button type="primary" :loading="generateLoading" @click="generateParam">生成</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type UploadUserFile } from 'element-plus'
import { Plus, MagicStick, Refresh } from '@element-plus/icons-vue'
import { ibcParamApi } from '@/api'
import type { IBCParamInfo } from '@/types'

const loading = ref(false)
const reloadLoading = ref(false)
const uploadLoading = ref(false)
const generateLoading = ref(false)

const ibcParams = ref<IBCParamInfo[]>([])

const showUploadDialog = ref(false)
const showGenerateDialog = ref(false)

const paramFiles = ref<UploadUserFile[]>([])

const uploadForm = ref({
  filename: '',
})

const generateForm = ref({
  districtName: 'tlcpchan.local',
  districtSerial: 1,
  years: 10,
})

onMounted(() => fetchIBCParams())

/**
 * 获取信任池列表
 */
async function fetchIBCParams() {
  loading.value = true
  try {
    ibcParams.value = await ibcParamApi.list()
  } catch (err) {
    console.error('获取 IBC 信任池列表失败', err)
  } finally {
    loading.value = false
  }
}

/**
 * 格式化日期
 * @param dateStr ISO8601 时间字符串
 */
function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

/**
 * 截断过长的主公钥指纹，便于表格展示
 * @param fingerprint 完整指纹（HEX）
 */
function truncateFingerprint(fingerprint: string): string {
  if (!fingerprint) return '-'
  if (fingerprint.length <= 20) {
    return fingerprint
  }
  return fingerprint.substring(0, 10) + '...' + fingerprint.substring(fingerprint.length - 10)
}

/**
 * 下载 KGC 公共参数
 * @param filename 信任池中的文件名
 */
async function download(filename: string) {
  try {
    await ibcParamApi.download(filename)
    ElMessage.success('公共参数下载成功')
  } catch (err: any) {
    ElMessage.error(err.message || '下载失败')
  }
}

/**
 * 上传 KGC 公共参数
 */
async function uploadParam() {
  if (!uploadForm.value.filename) {
    ElMessage.error('请输入文件名')
    return
  }
  if (paramFiles.value.length === 0) {
    ElMessage.error('请选择公共参数文件')
    return
  }

  uploadLoading.value = true
  try {
    await ibcParamApi.add(uploadForm.value.filename, paramFiles.value[0]!.raw as File)
    ElMessage.success('公共参数上传成功')
    showUploadDialog.value = false
    uploadForm.value.filename = ''
    paramFiles.value = []
    fetchIBCParams()
  } catch (err: any) {
    ElMessage.error(err.message || '上传失败')
  } finally {
    uploadLoading.value = false
  }
}

/**
 * 生成测试 KGC 公共参数
 */
async function generateParam() {
  if (!generateForm.value.districtName) {
    ElMessage.error('请填写 KGC 区域')
    return
  }

  generateLoading.value = true
  try {
    await ibcParamApi.generate(generateForm.value)
    ElMessage.success('测试 KGC 公共参数生成成功')
    showGenerateDialog.value = false
    fetchIBCParams()
  } catch (err: any) {
    ElMessage.error(err.message || '生成失败')
  } finally {
    generateLoading.value = false
  }
}

/**
 * 删除 KGC 公共参数（需输入文件名确认）
 * @param filename 信任池中的文件名
 */
async function remove(filename: string) {
  try {
    await ElMessageBox.prompt(
      `请输入文件名 <span style="color: red; font-weight: bold;">${filename}</span> 确认删除`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        dangerouslyUseHTMLString: true,
        inputPattern: new RegExp(`^${filename}$`),
        inputErrorMessage: '文件名不匹配，删除已取消',
      }
    )
    await ibcParamApi.delete(filename)
    ElMessage.success('公共参数已删除')
    fetchIBCParams()
  } catch (err) {
    if (err !== 'cancel') {
      console.error('删除失败', err)
    }
  }
}

/**
 * 重新扫描目录并重建信任池
 */
async function reloadPool() {
  reloadLoading.value = true
  try {
    await ibcParamApi.reload()
    ElMessage.success('信任池已重载，请重载相关实例使其生效')
    fetchIBCParams()
  } catch (err: any) {
    ElMessage.error(err.message || '重载失败')
  } finally {
    reloadLoading.value = false
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>