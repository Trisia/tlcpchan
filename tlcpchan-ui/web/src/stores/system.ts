import { defineStore } from 'pinia'
import { ref } from 'vue'
import { systemApi } from '@/api'
import type { SystemInfo, HealthInfo } from '@/types'

export const useSystemStore = defineStore('system', () => {
  const info = ref<SystemInfo | null>(null)
  const health = ref<HealthInfo | null>(null)
  const loading = ref(false)

  async function fetchInfo() {
    loading.value = true
    try {
      info.value = await systemApi.info()
    } finally {
      loading.value = false
    }
  }

  async function fetchHealth() {
    health.value = await systemApi.health()
  }

  return { info, health, loading, fetchInfo, fetchHealth }
})
