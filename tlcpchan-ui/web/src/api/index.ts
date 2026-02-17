import type { Instance, Certificate, SystemInfo, HealthInfo, VersionInfo, UIVersionInfo } from '@/types'

const API_BASE = '/api'

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!response.ok) {
    const text = await response.text()
    throw new Error(text || response.statusText)
  }
  if (response.status === 204) return undefined as T
  return response.json()
}

async function fetchVersion(): Promise<string> {
  try {
    const response = await fetch('/version.txt')
    if (response.ok) {
      return (await response.text()).trim()
    }
  } catch {
    // ignore
  }
  return 'dev'
}

export const instanceApi = {
  list: () => request<{ instances: Instance[]; total: number }>('/instances'),
  get: (name: string) => request<Instance>(`/instances/${name}`),
  create: (data: Partial<Instance>) =>
    request<{ name: string; status: string }>('/instances', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  update: (name: string, data: Partial<Instance>) =>
    request<{ name: string; status: string }>(`/instances/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  delete: (name: string) => request<void>(`/instances/${name}`, { method: 'DELETE' }),
  start: (name: string) => request<{ name: string; status: string }>(`/instances/${name}/start`, { method: 'POST' }),
  stop: (name: string) => request<{ name: string; status: string }>(`/instances/${name}/stop`, { method: 'POST' }),
  reload: (name: string) => request<{ name: string; status: string }>(`/instances/${name}/reload`, { method: 'POST' }),
  reloadCerts: (name: string) => request<{ name: string; status: string }>(`/instances/${name}/reload-certs`, { method: 'POST' }),
  stats: (name: string, period?: string) =>
    request<{ totalConnections: number; activeConnections: number; bytesReceived: number; bytesSent: number; avgLatencyMs: number }>(
      `/instances/${name}/stats${period ? `?period=${period}` : ''}`
    ),
  logs: (name: string, lines?: number, level?: string) =>
    request<{ logs: Array<{ time: string; level: string; message: string }> }>(
      `/instances/${name}/logs?${new URLSearchParams({ lines: String(lines || 100), ...(level && { level }) })}`
    ),
}



export const keyStoreApi = {
  list: () => request<{ keystores: Array<{
    name: string
    type: 'tlcp' | 'tls'
    keyParams: { algorithm: string; length: number; type: string }
    hasSignCert: boolean
    hasSignKey: boolean
    hasEncCert?: boolean
    hasEncKey?: boolean
    createdAt: string
    updatedAt: string
  }> }>('/keystores'),
  get: (name: string) => request<{
    name: string
    type: 'tlcp' | 'tls'
    keyParams: { algorithm: string; length: number; type: string }
    hasSignCert: boolean
    hasSignKey: boolean
    hasEncCert?: boolean
    hasEncKey?: boolean
    createdAt: string
    updatedAt: string
  }>(`/keystores/${name}`),
  create: async (data: {
    name: string
    type: 'tlcp' | 'tls'
    keyParams?: { algorithm: string; length: number }
    signCert: File
    signKey: File
    encCert?: File
    encKey?: File
  }) => {
    const formData = new FormData()
    formData.append('name', data.name)
    formData.append('type', data.type)
    if (data.keyParams) {
      formData.append('keyParams.algorithm', data.keyParams.algorithm)
      formData.append('keyParams.length', String(data.keyParams.length))
    }
    formData.append('signCert', data.signCert)
    formData.append('signKey', data.signKey)
    if (data.encCert) formData.append('encCert', data.encCert)
    if (data.encKey) formData.append('encKey', data.encKey)

    const response = await fetch(`${API_BASE}/keystores`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || response.statusText)
    }
    return response.json()
  },
  updateCertificates: async (name: string, data: {
    signCert?: File
    encCert?: File
  }) => {
    const formData = new FormData()
    if (data.signCert) formData.append('signCert', data.signCert)
    if (data.encCert) formData.append('encCert', data.encCert)

    const response = await fetch(`${API_BASE}/keystores/${name}/certificates`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || response.statusText)
    }
    return response.json()
  },
  delete: (name: string) => request<void>(`/keystores/${name}`, { method: 'DELETE' }),
  reload: (name: string) => request<{ message: string }>(`/keystores/${name}/reload`, { method: 'POST' }),
}

export const trustedApi = {
  list: () => request<Array<{
    name: string
    type: string
    serialNumber?: string
    subject?: string
    issuer?: string
    expiresAt?: string
    isCA?: boolean
  }>>('/trusted'),
  upload: async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch(`${API_BASE}/trusted`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || response.statusText)
    }
    return response.json()
  },
  delete: (name: string) => request<void>(`/trusted?name=${encodeURIComponent(name)}`, { method: 'DELETE' }),
  reload: () => request<{ message: string }>('/trusted/reload', { method: 'POST' }),
}

export const systemApi = {
  info: () => request<SystemInfo>('/system/info'),
  health: () => request<HealthInfo>('/system/health'),
  version: () => request<VersionInfo>('/system/version'),
}

export const configApi = {
  get: () => request<Record<string, unknown>>('/config'),
  reload: () => request<{ reloaded: boolean; changes: Record<string, unknown> }>('/config/reload', { method: 'POST' }),
}

export const uiApi = {
  version: async () => {
    const response = await fetch(`${API_BASE}/ui/version`, {
      headers: { 'Content-Type': 'application/json' },
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || response.statusText)
    }
    const result = await response.json()
    return result.data as UIVersionInfo
  },
  fetchStaticVersion: fetchVersion,
}
