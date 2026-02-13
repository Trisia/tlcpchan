import { defineStore } from 'pinia'
import { ref } from 'vue'
import { certificateApi } from '@/api'
import type { Certificate } from '@/types'

export const useCertificateStore = defineStore('certificate', () => {
  const certificates = ref<Certificate[]>([])
  const loading = ref(false)

  async function fetchCertificates() {
    loading.value = true
    try {
      const data = await certificateApi.list()
      certificates.value = data.certificates
    } finally {
      loading.value = false
    }
  }

  async function reloadCertificates() {
    await certificateApi.reload()
    await fetchCertificates()
  }

  return { certificates, loading, fetchCertificates, reloadCertificates }
})
