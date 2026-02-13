import { defineStore } from 'pinia'
import { ref } from 'vue'
import { instanceApi } from '@/api'
import type { Instance } from '@/types'

export const useInstanceStore = defineStore('instance', () => {
  const instances = ref<Instance[]>([])
  const loading = ref(false)

  async function fetchInstances() {
    loading.value = true
    try {
      const data = await instanceApi.list()
      instances.value = data.instances
    } finally {
      loading.value = false
    }
  }

  async function startInstance(name: string) {
    await instanceApi.start(name)
    await fetchInstances()
  }

  async function stopInstance(name: string) {
    await instanceApi.stop(name)
    await fetchInstances()
  }

  async function deleteInstance(name: string) {
    await instanceApi.delete(name)
    await fetchInstances()
  }

  return { instances, loading, fetchInstances, startInstance, stopInstance, deleteInstance }
})
