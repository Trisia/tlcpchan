import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'
import type { SystemInfo, HealthInfo } from '@/types'

const API_BASE = '/api/v1'

export const useSystemStore = defineStore('system', () => {
  const info = ref<SystemInfo | null>(null)
  const health = ref<HealthInfo | null>(null)
  const loading = ref(false)

  function fetchInfo() {
    loading.value = true
    axios.get(`${API_BASE}/system/info`)
      .then((res) => {
        info.value = res.data
      })
      .catch((err) => {
        console.error('获取系统信息失败', err)
      })
      .finally(() => {
        loading.value = false
      })
  }

  function fetchHealth() {
    axios.get(`${API_BASE}/system/health`)
      .then((res) => {
        health.value = res.data
      })
      .catch((err) => {
        console.error('获取健康状态失败', err)
      })
  }

  return { info, health, loading, fetchInfo, fetchHealth }
})
