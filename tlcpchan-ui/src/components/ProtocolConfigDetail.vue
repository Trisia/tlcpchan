<template>
  <div class="protocol-config-detail">
    <!-- 分组一：PKI 配置（X.509 证书身份，TLCP 与 TLS 通用） -->
    <section class="config-group">
      <div class="config-group-title">PKI 配置</div>
      <el-descriptions :column="1" border size="small">
        <!-- 未配置证书 keystore 时给出明确说明，避免出现空白表格 -->
        <el-descriptions-item v-if="!config.keystore" label="证书 keystore">未配置</el-descriptions-item>
        <el-descriptions-item v-if="config.keystore" label="Keystore 类型">
          <el-tag :type="config.keystore.type === 'named' ? 'primary' : 'success'" size="small">
            {{ config.keystore.type === 'named' ? '引用已有密钥 (named)' : '直接指定文件 (file)' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="config.keystore && config.keystore.type === 'named'" label="Keystore 名称">
          {{ config.keystore.name || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="config.keystore && config.keystore.type === 'file'" label="签名证书路径">
          {{ config.keystore.params?.['sign-cert'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="config.keystore && config.keystore.type === 'file'" label="签名密钥路径">
          {{ config.keystore.params?.['sign-key'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="config.keystore && config.keystore.type === 'file' && isTlcp" label="加密证书路径">
          {{ config.keystore.params?.['enc-cert'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="config.keystore && config.keystore.type === 'file' && isTlcp" label="加密密钥路径">
          {{ config.keystore.params?.['enc-key'] || '-' }}
        </el-descriptions-item>
      </el-descriptions>
    </section>

    <!-- 分组二：IBC 配置（SM9 标识身份，仅 TLCP 侧存在） -->
    <section class="config-group" v-if="isTlcp">
      <div class="config-group-title">IBC 配置</div>
      <el-descriptions :column="1" border size="small">
        <!-- 未配置 IBC 身份时给出明确说明，避免出现空白表格 -->
        <el-descriptions-item v-if="!tlcpConfig.ibcKeystore" label="IBC 身份 keystore">未配置</el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore" label="IBC 身份 keystore 类型">
          <el-tag :type="tlcpConfig.ibcKeystore.type === 'named' ? 'primary' : 'warning'" size="small">
            {{ tlcpConfig.ibcKeystore.type === 'named' ? '引用已有密钥 (named)' : '直接指定 IBC 材料文件 (ibc-file)' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'named'" label="IBC 身份名称">
          {{ tlcpConfig.ibcKeystore.name || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'ibc-file'" label="IBC 标识文件">
          {{ tlcpConfig.ibcKeystore.params?.['identity'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'ibc-file'" label="IBC 公共参数文件">
          {{ tlcpConfig.ibcKeystore.params?.['params'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'ibc-file'" label="IBC 签名私钥文件">
          {{ tlcpConfig.ibcKeystore.params?.['sign-key'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'ibc-file'" label="IBC 加密私钥文件">
          {{ tlcpConfig.ibcKeystore.params?.['enc-key'] || '-' }}
        </el-descriptions-item>
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'ibc-file'" label="IBC 密钥交换私钥文件">
          {{ tlcpConfig.ibcKeystore.params?.['kex-key'] || '-' }}
        </el-descriptions-item>
        <!--
          IBC 身份元信息：named 引用时按密钥库名称查询补齐；ibc-file 直接指定材料时
          配置中只有文件路径，标识与私钥用途以文件内容为准。
        -->
        <el-descriptions-item v-if="tlcpConfig.ibcKeystore?.type === 'ibc-file'" label="IBC 身份元信息">
          <el-tag type="info" size="small">直接指定材料文件，标识、KGC 区域与私钥用途以文件内容为准</el-tag>
        </el-descriptions-item>
        <template v-if="tlcpConfig.ibcKeystore?.type === 'named'">
          <el-descriptions-item v-if="ibcInfoLoading" label="IBC 身份元信息">加载中…</el-descriptions-item>
          <el-descriptions-item v-else-if="ibcInfoError" label="IBC 身份元信息">
            <el-tag type="danger" size="small">{{ ibcInfoError }}</el-tag>
          </el-descriptions-item>
          <template v-else>
            <el-descriptions-item label="IBC 标识">{{ ibcInfo?.identity || '-' }}</el-descriptions-item>
            <el-descriptions-item label="KGC 区域">{{ ibcInfo?.districtName || '-' }}</el-descriptions-item>
            <!-- 未提供 KGC 公共参数时后端返回序号 0，此时按“-”展示而非 0 -->
            <el-descriptions-item label="KGC 序号">{{ ibcInfo?.districtName ? ibcInfo?.districtSerial : '-' }}</el-descriptions-item>
            <el-descriptions-item label="KGC 公共参数">
              <el-tag :type="ibcInfo?.hasParams ? 'success' : 'info'" size="small">
                {{ ibcInfo?.hasParams ? '已提供' : '未提供' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="公共参数有效期">
              <span v-if="ibcInfo?.hasParams">
                {{ formatDateTime(ibcInfo?.notBefore, '不限') }} ~ {{ formatDateTime(ibcInfo?.notAfter, '不限') }}
              </span>
              <span v-else>-</span>
            </el-descriptions-item>
            <el-descriptions-item label="签名私钥 (hid=0x01)">
              <el-tag :type="ibcInfo?.hasSignKey ? 'success' : 'info'" size="small">
                {{ ibcInfo?.hasSignKey ? '已提供' : '未提供' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="加密私钥 (hid=0x03)">
              <el-tag :type="ibcInfo?.hasEncryptKey ? 'success' : 'info'" size="small">
                {{ ibcInfo?.hasEncryptKey ? '已提供' : '未提供' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="密钥交换私钥 (hid=0x02)">
              <el-tag :type="ibcInfo?.hasKeyExchangeKey ? 'success' : 'info'" size="small">
                {{ ibcInfo?.hasKeyExchangeKey ? '已提供' : '未提供' }}
              </el-tag>
            </el-descriptions-item>
          </template>
        </template>
      </el-descriptions>
    </section>

    <!-- 分组三：协议参数（TLCP 侧含 IBC/IBSDH 套件勾选情况） -->
    <section class="config-group">
      <div class="config-group-title">{{ isTlcp ? 'TLCP 协议配置' : 'TLS 协议配置' }}</div>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="客户端认证类型">
          {{ formatClientAuth(config.clientAuthType) }}
        </el-descriptions-item>
        <el-descriptions-item label="最低版本">
          {{ config.minVersion || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="最高版本">
          {{ config.maxVersion || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="密码套件">
          <el-tag
            v-for="(cs, i) in (config.cipherSuites || [])"
            :key="i"
            size="small"
            :type="isIBCSuite(cs) ? 'warning' : 'primary'"
            style="margin-right: 4px; margin-bottom: 4px"
          >
            {{ cs }}
          </el-tag>
          <span v-if="!config.cipherSuites?.length">-</span>
        </el-descriptions-item>
        <el-descriptions-item label="IBC/IBSDH 套件">
          <template v-if="ibcSuites.length">
            <el-tag
              v-for="cs in ibcSuites"
              :key="cs"
              type="warning"
              size="small"
              style="margin-right: 4px; margin-bottom: 4px"
            >
              {{ cs }}
            </el-tag>
          </template>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="会话票据" v-if="!isTlcp">
          <el-tag :type="(config as TLSConfig).sessionTickets ? 'success' : 'info'" size="small">
            {{ (config as TLSConfig).sessionTickets ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="握手重用">
          <el-tag :type="config.sessionCache ? 'success' : 'info'" size="small">
            {{ config.sessionCache ? '启用' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="跳过证书验证">
          <el-tag :type="config.insecureSkipVerify ? 'danger' : 'success'" size="small">
            {{ config.insecureSkipVerify ? '启用（不安全）' : '禁用' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { formatClientAuth } from '@/constants/clientAuth'
import { isIBCSuite } from '@/constants/cipherSuite'
import { formatDateTime } from '@/constants/datetime'
import { keyStoreApi } from '@/api'
import type { IBCKeyStoreInfo, TLCPConfig, TLSConfig } from '@/types'

interface Props {
  config: TLCPConfig | TLSConfig
  isTlcp: boolean
}

const props = defineProps<Props>()

// IBC 身份仅存在于 TLCP 配置中
const tlcpConfig = computed(() => props.config as TLCPConfig)

// 已勾选的 IBC/IBSDH 套件
const ibcSuites = computed(() => (props.config.cipherSuites || []).filter((cs) => isIBCSuite(cs)))

// 引用的 IBC 身份密钥库只读元信息（仅 isTlcp 且 ibcKeystore.type=named 时查询）
const ibcInfo = ref<IBCKeyStoreInfo | null>(null)
// 元信息加载中标志
const ibcInfoLoading = ref(false)
// 查询失败或密钥库缺少 IBC 元信息时的提示文案
const ibcInfoError = ref('')

// named 引用时按密钥库名称查询只读元信息；ibc-file 直接指定材料时不查询
watch(
  () => (props.isTlcp ? tlcpConfig.value.ibcKeystore : undefined),
  async (ibcKeystore) => {
    ibcInfo.value = null
    ibcInfoError.value = ''
    if (ibcKeystore?.type !== 'named' || !ibcKeystore.name) return

    ibcInfoLoading.value = true
    try {
      const detail = await keyStoreApi.get(ibcKeystore.name)
      ibcInfo.value = detail?.ibc || null
      if (!ibcInfo.value) ibcInfoError.value = '该密钥库不含 IBC 元信息'
    } catch (err) {
      console.error('获取 IBC 身份元信息失败:', err)
      ibcInfoError.value = '无法获取（密钥库可能已被删除）'
    } finally {
      ibcInfoLoading.value = false
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.protocol-config-detail {
  margin-bottom: 16px;
}
/* 分组区块之间留出间距，视觉上区分 PKI / IBC / 协议三类配置 */
.config-group + .config-group {
  margin-top: 16px;
}
/* 分组标题：左侧色条 + 加粗文字，比折叠面板的子标题更醒目 */
.config-group-title {
  margin-bottom: 8px;
  padding-left: 8px;
  border-left: 3px solid var(--el-color-primary);
  font-size: 14px;
  font-weight: 600;
  line-height: 16px;
  color: var(--el-text-color-primary);
}
</style>