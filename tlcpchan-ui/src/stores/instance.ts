import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getInstances, startInstance as apiStartInstance, stopInstance as apiStopInstance } from '@/utils/http'
import type { Instance } from '@/types'

export const useInstanceStore = defineStore('instance', () => {
  const instances = ref<Instance[]>([])
  const loading = ref(false)

  function setInstances(data: Instance[]) {
    instances.value = data
  }

  function updateInstance(name: string, updates: Partial<Instance>) {
    const index = instances.value.findIndex(inst => inst.name === name)
    if (index !== -1) {
      instances.value[index] = { ...instances.value[index], ...updates } as Instance
    }
  }

  async function fetchInstances() {
    loading.value = true
    try {
      const data = await getInstances()
      instances.value = data
    } finally {
      loading.value = false
    }
  }

  async function startInstance(name: string) {
    await apiStartInstance(name)
    await fetchInstances()
  }

  async function stopInstance(name: string) {
    await apiStopInstance(name)
    await fetchInstances()
  }

  return { 
    instances, 
    loading,
    setInstances, 
    updateInstance,
    fetchInstances,
    startInstance,
    stopInstance
  }
})