import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/utils/http'
import type { SystemInfo, HealthStatus } from '@/types'

export const useSystemStore = defineStore('system', () => {
  const info = ref<SystemInfo | null>(null)
  const health = ref<HealthStatus | null>(null)

  function setInfo(data: SystemInfo | null) {
    info.value = data
  }

  function setHealth(data: HealthStatus | null) {
    health.value = data
  }

  async function fetchInfo() {
    try {
      const data = await http.get('/system/info')
      info.value = data
    } catch (error) {
      console.error('获取系统信息失败:', error)
    }
  }

  async function fetchHealth() {
    try {
      const data = await http.get('/health')
      health.value = data
    } catch (error) {
      console.error('获取健康状态失败:', error)
    }
  }

  return { 
    info, 
    health, 
    setInfo, 
    setHealth,
    fetchInfo,
    fetchHealth
  }
})
