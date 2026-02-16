import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'
import type { Certificate } from '@/types'

const API_BASE = '/api/v1'

export const useCertificateStore = defineStore('certificate', () => {
  const certificates = ref<Certificate[]>([])
  const loading = ref(false)

  function fetchCertificates() {
    loading.value = true
    axios.get(`${API_BASE}/certificates`)
      .then((res) => {
        certificates.value = res.data.certificates
      })
      .catch((err) => {
        console.error('获取证书列表失败', err)
      })
      .finally(() => {
        loading.value = false
      })
  }

  function reloadCertificates() {
    axios.post(`${API_BASE}/certificates/reload`)
      .then(() => {
        fetchCertificates()
      })
      .catch((err) => {
        console.error('重载证书失败', err)
      })
  }

  return { certificates, loading, fetchCertificates, reloadCertificates }
})
