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
                <el-radio :value="CertType.IBC">标识密码 (IBC)</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>

        <!-- IBC(SM9) 身份：选择信任池中的 KGC 区域 + 本端标识，一次性派生三把用户私钥 -->
        <template v-if="generateForm.type === CertType.IBC">
          <el-divider content-position="left">IBC 身份 (SM9)</el-divider>
          <el-alert
            title="将使用服务端内置的测试 KGC 主密钥，一次性派生三把用户私钥：签名私钥 (hid=0x01)、加密私钥 (hid=0x03)、密钥交换私钥 (hid=0x02)。生产环境应由外部 KGC 派生后导入。"
            type="warning"
            :closable="false"
            style="margin-bottom: 20px;"
          />
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="KGC 区域" required>
                <el-select
                  v-model="selectedKGC"
                  placeholder="请选择信任池中的 KGC 公共参数"
                  style="width: 100%;"
                  @change="onKGCChange"
                >
                  <el-option
                    v-for="p in ibcParams"
                    :key="p.filename"
                    :label="`${p.districtName}#${p.districtSerial} (${p.filename})`"
                    :value="p.filename"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="标识" required>
                <el-input v-model="generateForm.identity" placeholder="server@tlcpchan.local" />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="区域名称">
                <el-input v-model="generateForm.districtName" placeholder="选择 KGC 后自动填充" disabled />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="KGC 序号">
                <el-input-number v-model="generateForm.districtSerial" :min="0" :max="65535" style="width: 100%;" disabled />
              </el-form-item>
            </el-col>
          </el-row>
          <el-alert
            v-if="ibcParams.length === 0"
            title="IBC 信任池为空，请先在「IBC 信任池」页面添加或生成 KGC 公共参数"
            type="error"
            :closable="false"
            style="margin-bottom: 20px;"
          />
        </template>

        <template v-if="generateForm.type !== CertType.IBC">
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
        </template>
      </el-form>
      
      <template #footer>
        <el-button @click="goBack">取消</el-button>
        <el-button type="primary" :loading="generateLoading" @click="generateKeyStore">生成</el-button>
      </template>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { keyStoreApi, ibcParamApi } from '@/api'
import { CertType } from '@/types'
import type { IBCParamInfo } from '@/types'
import { keystoreListLocation } from '@/constants/keystoreTab'

const route = useRoute()
const router = useRouter()
const generateLoading = ref(false)

// IBC 信任池列表，供选择 KGC 公共参数所属区域
const ibcParams = ref<IBCParamInfo[]>([])
// 选中的信任池条目文件名
const selectedKGC = ref('')

const generateForm = ref({
  name: '',
  type: CertType.TLCP as 'tlcp' | 'tls' | 'ibc',
  protected: false as boolean,
  // 仅 IBC 类型使用
  identity: '',
  districtName: '',
  districtSerial: 0,
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

onMounted(() => {
  // 由密钥管理页面跳转而来时通过 query.type 预设密钥类型（白名单校验，非法或缺失时保持默认 TLCP）
  const presetType = route.query.type
  if (presetType === CertType.TLCP || presetType === CertType.TLS || presetType === CertType.IBC) {
    generateForm.value.type = presetType
  }
  loadIBCParams()
})

/**
 * 加载 IBC 信任池列表（生成 IBC 身份时选择 KGC 区域）
 */
async function loadIBCParams() {
  try {
    ibcParams.value = await ibcParamApi.list()
  } catch (err) {
    console.error('获取 IBC 信任池列表失败:', err)
  }
}

/**
 * 选择信任池条目后回填 KGC 区域名与序号
 * @param filename 信任池条目文件名
 */
function onKGCChange(filename: string) {
  const param = ibcParams.value.find((p) => p.filename === filename)
  if (param) {
    generateForm.value.districtName = param.districtName
    generateForm.value.districtSerial = param.districtSerial
  }
}

/**
 * 返回密钥列表页，并停留在当前密钥类型所属的分类 Tab
 */
function goBack() {
  router.push(keystoreListLocation(generateForm.value.type))
}

/**
 * 生成密钥存储
 * IBC 类型提交 {name, type, protected, identity, districtName, districtSerial}；
 * 其余类型沿用证书生成流程。
 */
async function generateKeyStore() {
  if (!generateForm.value.name) {
    ElMessage.error('请填写密钥名称')
    return
  }

  if (generateForm.value.type === CertType.IBC) {
    if (!generateForm.value.districtName) {
      ElMessage.error('请选择 KGC 公共参数所属区域')
      return
    }
    if (!generateForm.value.identity) {
      ElMessage.error('请填写 IBC 标识')
      return
    }

    generateLoading.value = true
    try {
      await keyStoreApi.generate({
        name: generateForm.value.name,
        type: CertType.IBC,
        protected: generateForm.value.protected,
        identity: generateForm.value.identity,
        districtName: generateForm.value.districtName,
        districtSerial: generateForm.value.districtSerial,
      })
      ElMessage.success('IBC 身份密钥生成成功')
      goBack()
    } catch (err: any) {
      ElMessage.error(err.message || '生成失败')
    } finally {
      generateLoading.value = false
    }
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
