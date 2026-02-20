import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RootCertInfo } from '@/types'

export const useCertificateStore = defineStore('certificate', () => {
  const certificates = ref<RootCertInfo[]>([])

  function setCertificates(data: RootCertInfo[]) {
    certificates.value = data
  }

  return { certificates, setCertificates }
})
