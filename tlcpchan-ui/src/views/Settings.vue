<template>
  <div class="settings">
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>API 配置</span>
          </template>
          <el-form label-width="120px">
            <el-form-item label="监听地址">
              <el-input v-model="config.server.api.address" placeholder=":20080" />
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span>MCP 配置</span>
          </template>
          <el-form label-width="120px">
            <el-form-item label="启用 MCP">
              <el-switch v-model="config.mcp!.enabled" />
            </el-form-item>
            <el-form-item label="API 密钥" v-if="config.mcp!.enabled">
              <el-input v-model="config.mcp!.apiKey" type="password" show-password placeholder="留空表示无需认证" />
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>日志配置</span>
      </template>
      <el-form label-width="140px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="启用日志">
              <el-switch v-model="logConfig!.enabled" />
            </el-form-item>
            <el-form-item label="日志级别">
              <el-select v-model="logConfig!.level">
              <el-option value="debug" label="debug" />
              <el-option value="info" label="info" />
              <el-option value="warn" label="warn" />
              <el-option value="error" label="error" />
              </el-select>
            </el-form-item>
            <el-form-item label="日志文件路径">
              <el-input v-model="logConfig!.file" placeholder="./logs/tlcpchan.log" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大文件大小 (MB)">
              <el-input-number v-model="logConfig!.maxSize" :min="1" />
            </el-form-item>
            <el-form-item label="最大备份文件数">
              <el-input-number v-model="logConfig!.maxBackups" :min="0" />
            </el-form-item>
            <el-form-item label="最大保留天数">
              <el-input-number v-model="logConfig!.maxAge" :min="0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="压缩旧日志">
          <el-switch v-model="logConfig!.compress" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <span>操作</span>
      </template>
      <el-button type="primary" @click="saveConfig" :loading="saving">
        <el-icon><DocumentChecked /></el-icon>
        保存配置
      </el-button>
      <el-button type="success" @click="reloadConfig" :loading="reloading">
        <el-icon><Refresh /></el-icon>
        重载配置
      </el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { configApi } from '@/api'
import type { Config } from '@/types'

const config = ref<Config>({
  server: {
    api: { address: ':20080' },
    log: {
      level: 'info',
      file: './logs/tlcpchan.log',
      maxSize: 100,
      maxBackups: 5,
      maxAge: 30,
      compress: true,
      enabled: true
    }
  },
  mcp: {
    enabled: false,
    apiKey: ''
  }
})

const saving = ref(false)
const reloading = ref(false)

const logConfig = computed({
  get: () => {
    if (!config.value.server.log) {
      config.value.server.log = {
        level: 'info',
        file: './logs/tlcpchan.log',
        maxSize: 100,
        maxBackups: 5,
        maxAge: 30,
        compress: true,
        enabled: true
      }
    }
    return config.value.server.log
  },
  set: (val) => {
    config.value.server.log = val
  }
})

onMounted(() => {
  fetchConfig()
})

async function fetchConfig() {
  try {
    const data = await configApi.get()
    config.value = data
  } catch (error) {
    console.error('获取配置失败:', error)
    ElMessage.error('获取配置失败')
  }
}

async function saveConfig() {
  try {
    saving.value = true
    await configApi.update(config.value)
    ElMessage.success('配置已保存')
  } catch (error: any) {
    console.error('保存配置失败:', error)
    ElMessage.error(error.response?.data || '保存配置失败')
  } finally {
    saving.value = false
  }
}

async function reloadConfig() {
  try {
    reloading.value = true
    await configApi.reload()
    ElMessage.success('配置已重载')
    await fetchConfig()
  } catch (error) {
    console.error('重载配置失败:', error)
    ElMessage.error('重载配置失败')
  } finally {
    reloading.value = false
  }
}
</script>
