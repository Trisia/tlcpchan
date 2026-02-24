import axios from 'axios'
import type {
  GenerateKeyStoreRequest,
  GenerateRootCARequest,
} from '@/types'

export const API_CONFIG = {
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: parseInt(import.meta.env.VITE_API_TIMEOUT || '10000'),
  headers: {
    'Content-Type': 'application/json',
  },
}

export const http = axios.create({
  baseURL: API_CONFIG.baseURL,
  timeout: API_CONFIG.timeout,
  headers: API_CONFIG.headers,
})

export const keyStoreApi = {
  list: async () => {
    const res = await http.get('/security/keystores')
    return { keystores: res.data || [] }
  },

  get: async (name: string) => {
    const res = await http.get(`/security/keystores/${name}`)
    return res.data
  },

  create: async (data: any) => {
    const formData = new FormData()
    formData.append('name', data.name)
    if (data.type) formData.append('type', data.type)
    if (data.signCert) formData.append('signCert', data.signCert)
    if (data.signKey) formData.append('signKey', data.signKey)
    if (data.encCert) formData.append('encCert', data.encCert)
    if (data.encKey) formData.append('encKey', data.encKey)
    
    const res = await http.post('/security/keystores', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data
  },

  generate: async (data: GenerateKeyStoreRequest) => {
    const res = await http.post('/security/keystores/generate', data)
    return res.data
  },

  delete: async (name: string) => {
    await http.delete(`/security/keystores/${name}`)
  },

  exportCSR: async (name: string, data: {
    keyType: 'sign' | 'enc'
    csrParams: {
      commonName: string
      country?: string
      stateOrProvince?: string
      locality?: string
      org?: string
      orgUnit?: string
      emailAddress?: string
      dnsNames?: string[]
      ipAddresses?: string[]
    }
  }) => {
    const res = await http.post(`/security/keystores/${name}/export-csr`, data, {
      responseType: 'blob'
    })
    
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    let filename = `${name}-csr-${Date.now()}.csr`
    const contentDisposition = res.headers['content-disposition']
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="([^"]+)/)
      if (match) {
        filename = match[1]
      }
    }
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },

  updateCertificates: async (name: string, data: any) => {
    const formData = new FormData()
    if (data.signCert) formData.append('signCert', data.signCert)
    if (data.encCert) formData.append('encCert', data.encCert)
    
    const res = await http.post(`/security/keystores/${name}`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data
  },

  update: async (name: string, data: { params: Record<string, string> }) => {
    const res = await http.put(`/security/keystores/${name}`, data)
    return res.data
  },

  getInstances: async (name: string) => {
    const res = await http.get(`/security/keystores/${name}/instances`)
    return res.data
  },
}

export const rootCertApi = {
  list: async () => {
    const res = await http.get('/security/rootcerts')
    return res.data || []
  },

  download: async (filename: string) => {
    const res = await http.get(`/security/rootcerts/${filename}`, {
      responseType: 'blob'
    })

    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },

  add: async (filename: string, file: File) => {
    const formData = new FormData()
    formData.append('filename', filename)
    formData.append('cert', file)
    const res = await http.post('/security/rootcerts', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data
  },

  generate: async (data: GenerateRootCARequest) => {
    const res = await http.post('/security/rootcerts/generate', data)
    return res.data
  },

  delete: async (filename: string) => {
    await http.delete(`/security/rootcerts/${filename}`)
  },

  reload: async () => {
    await http.post('/security/rootcerts/reload')
  },
}

export const trustedApi = {
  list: rootCertApi.list,
  download: rootCertApi.download,
  upload: rootCertApi.add,
  delete: rootCertApi.delete,
}

export const instanceApi = {
  list: async () => {
    const res = await http.get('/instances')
    return res.data || []
  },

  get: async (name: string) => {
    const res = await http.get(`/instances/${name}`)
    return res.data
  },

  create: async (data: any) => {
    const res = await http.post('/instances', data)
    return res.data
  },

  update: async (name: string, data: any) => {
    const res = await http.put(`/instances/${name}`, data)
    return res.data
  },

  delete: async (name: string) => {
    await http.delete(`/instances/${name}`)
  },

  start: async (name: string) => {
    const res = await http.post(`/instances/${name}/start`)
    return res.data
  },

  stop: async (name: string) => {
    const res = await http.post(`/instances/${name}/stop`)
    return res.data
  },

  reload: async (name: string) => {
    const res = await http.post(`/instances/${name}/reload`)
    return res.data
  },

  restart: async (name: string) => {
    const res = await http.post(`/instances/${name}/restart`)
    return res.data
  },

  stats: async (name: string, period?: string) => {
    const params = period ? { period } : {}
    const res = await http.get(`/instances/${name}/stats`, { params })
    return res.data
  },

  logs: async (name: string, params?: { lines?: number; level?: string }) => {
    const res = await http.get(`/instances/${name}/logs`, { params })
    return res.data || []
  },

  health: async (name: string, params?: { timeout?: number; protocol?: string }) => {
    const res = await http.get(`/instances/${name}/health`, { params })
    return res.data
  },
}

export const systemApi = {
  info: async () => {
    const res = await http.get('/system/info')
    return res.data
  },

  health: async () => {
    const res = await http.get('/system/health')
    return res.data
  },

  version: async () => {
    const res = await http.get('/system/version')
    return res.data
  },
}

export const logsApi = {
  list: async () => {
    const res = await http.get('/system/logs')
    return res.data
  },

  content: async (params?: { file?: string; lines?: number; level?: string }) => {
    const res = await http.get('/system/logs/content', { params })
    return res.data
  },

  download: async (filename: string) => {
    const res = await http.get(`/system/logs/download/${filename}`, {
      responseType: 'blob'
    })
    
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },

  downloadAll: async () => {
    const res = await http.get('/system/logs/download-all', {
      responseType: 'blob'
    })
    
    const contentDisposition = res.headers['content-disposition']
    let filename = 'tlcpchan-logs.zip'
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="?([^"]+)"?/)
      if (match) {
        filename = match[1]
      }
    }
    
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },
}

export const configApi = {
  get: async () => {
    const res = await http.get('/config')
    return res.data
  },

  update: async (data: any) => {
    const res = await http.post('/config', data)
    return res.data
  },

  reload: async () => {
    const res = await http.post('/config/reload')
    return res.data
  },

  validate: async (path?: string) => {
    const res = await http.post('/config/validate', path ? { path } : {})
    return res.data
  },
}

export default {
  keyStoreApi,
  rootCertApi,
  trustedApi,
  instanceApi,
  systemApi,
  logsApi,
  configApi,
  http,
}
