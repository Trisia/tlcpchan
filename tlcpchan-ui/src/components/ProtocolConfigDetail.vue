<template>
  <div class="protocol-config-detail">
    <el-descriptions :column="1" border size="small">
      <el-descriptions-item label="客户端认证类型">
        {{ config.clientAuthType || 'no-client-cert' }}
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
          style="margin-right: 4px; margin-bottom: 4px"
        >
          {{ cs }}
        </el-tag>
        <span v-if="!config.cipherSuites?.length">-</span>
      </el-descriptions-item>
      <el-descriptions-item label="会话票据">
        <el-tag :type="config.sessionTickets ? 'success' : 'info'" size="small">
          {{ config.sessionTickets ? '启用' : '禁用' }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="会话缓存">
        <el-tag :type="config.sessionCache ? 'success' : 'info'" size="small">
          {{ config.sessionCache ? '启用' : '禁用' }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="跳过证书验证">
        <el-tag :type="config.insecureSkipVerify ? 'danger' : 'success'" size="small">
          {{ config.insecureSkipVerify ? '启用（不安全）' : '禁用' }}
        </el-tag>
      </el-descriptions-item>
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
  </div>
</template>

<script setup lang="ts">
import type { TLCPConfig, TLSConfig } from '@/types'

interface Props {
  config: TLCPConfig | TLSConfig
  isTlcp: boolean
}

defineProps<Props>()
</script>

<style scoped>
.protocol-config-detail {
  margin-bottom: 16px;
}
</style>
