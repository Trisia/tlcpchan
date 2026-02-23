<template>
  <div class="settings">
    <el-card class="system-info-card">
      <template #header>
        <span>系统信息</span>
      </template>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="UI版本">{{ uiVersion }}</el-descriptions-item>
        <el-descriptions-item label="后端版本">{{ health?.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="系统">{{ info?.os }}/{{ info?.arch }}</el-descriptions-item>
        <el-descriptions-item label="CPU核心数">{{ info?.numCpu }}</el-descriptions-item>
        <el-descriptions-item label="Goroutines">{{ info?.numGoroutine }}</el-descriptions-item>
        <el-descriptions-item label="内存">{{ info?.memAllocMb }} MB / {{ info?.memSysMb }} MB</el-descriptions-item>
        <el-descriptions-item label="运行时长">{{ info?.uptime || '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-row :gutter="20" style="margin-top: 20px">
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
            <el-form-item label="API 密钥">
              <el-input v-model="config.mcp!.apiKey" type="password" show-password placeholder="留空表示无需认证" />
            </el-form-item>
            <el-form-item label="对接地址">
              <el-input :value="mcpConnectUrl" readonly />
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
        <el-icon>
          <DocumentChecked />
        </el-icon>
        保存配置
      </el-button>
      <el-button type="success" @click="reloadConfig" :loading="reloading">
        <el-icon>
          <Refresh />
        </el-icon>
        重载配置
      </el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { configApi, systemApi } from '@/api'
import type { Config, SystemInfo, HealthStatus } from '@/types'
import { DocumentChecked, Refresh } from '@element-plus/icons-vue'
import axios from 'axios'

const info = ref<SystemInfo | null>(null)
const health = ref<HealthStatus | null>(null)
const uiVersion = ref('dev')

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
    apiKey: ''
  },
  keystores: [],
  instances: []
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

const mcpConnectUrl = computed(() => {
  const host = window.location.host
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const apiKey = config.value.mcp?.apiKey
  let url = `${protocol}//${host}/api/mcp/ws`
  if (apiKey) {
    url += `?api_key=${apiKey}`
  }
  return url
})

onMounted(() => {
  // 获取UI版本
  axios.get('./version.txt', { responseType: 'text' })
    .then((response) => {
      uiVersion.value = response.data.trim()
    })
    .catch(() => {
      uiVersion.value = 'dev'
    })

  Promise.all([fetchConfig(), fetchInfo(), fetchHealth()])
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

async function fetchInfo() {
  try {
    info.value = await systemApi.info()
  } catch (error) {
    console.error('获取系统信息失败:', error)
  }
}

async function fetchHealth() {
  try {
    health.value = await systemApi.health()
  } catch (error) {
    console.error('获取健康状态失败:', error)
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
