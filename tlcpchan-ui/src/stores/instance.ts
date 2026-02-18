import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'
import type { Instance } from '@/types'

const API_BASE = '/api/v1'

export const useInstanceStore = defineStore('instance', () => {
  const instances = ref<Instance[]>([])
  const loading = ref(false)

  function fetchInstances() {
    loading.value = true
    axios.get(`${API_BASE}/instances`)
      .then((res) => {
        instances.value = res.data.instances
      })
      .catch((err) => {
        console.error('获取实例列表失败', err)
      })
      .finally(() => {
        loading.value = false
      })
  }

  function startInstance(name: string) {
    axios.post(`${API_BASE}/instances/${name}/start`)
      .then(() => {
        fetchInstances()
      })
      .catch((err) => {
        console.error('启动实例失败', err)
      })
  }

  function stopInstance(name: string) {
    axios.post(`${API_BASE}/instances/${name}/stop`)
      .then(() => {
        fetchInstances()
      })
      .catch((err) => {
        console.error('停止实例失败', err)
      })
  }

  function deleteInstance(name: string) {
    axios.delete(`${API_BASE}/instances/${name}`)
      .then(() => {
        fetchInstances()
      })
      .catch((err) => {
        console.error('删除实例失败', err)
      })
  }

  return { instances, loading, fetchInstances, startInstance, stopInstance, deleteInstance }
})
