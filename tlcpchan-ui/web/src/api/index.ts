import type {
  Instance,
  InstanceConfig,
  InstanceStats,
  KeyStoreInfo,
  GenerateKeyStoreRequest,
  RootCertInfo,
  GenerateRootCARequest,
  SystemInfo,
  HealthStatus,
  VersionInfo,
  Config,
  InstanceHealthResponse,
} from '@/types'

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
  }
  return 'dev'
}

export const instanceApi = {
  list: () => request<Instance[]>('/instances'),
  get: (name: string) => request<Instance>(`/instances/${name}`),
  create: (data: Partial<InstanceConfig>) =>
    request<{ name: string; status: string }>('/instances', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  update: (name: string, data: Partial<InstanceConfig>) =>
    request<{ name: string; status: string }>(`/instances/${name}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  delete: (name: string) => request<void>(`/instances/${name}`, { method: 'DELETE' }),
  start: (name: string) => request<{ status: string }>(`/instances/${name}/start`, { method: 'POST' }),
  stop: (name: string) => request<{ status: string }>(`/instances/${name}/stop`, { method: 'POST' }),
  reload: (name: string) => request<{ status: string }>(`/instances/${name}/reload`, { method: 'POST' }),
  restart: (name: string) => request<{ status: string }>(`/instances/${name}/restart`, { method: 'POST' }),
  stats: (name: string, period?: string) =>
    request<InstanceStats>(`/instances/${name}/stats${period ? `?period=${period}` : ''}`),
  logs: (name: string, lines?: number, level?: string) =>
    request<Array<{ timestamp: string; level: string; message: string }>>(
      `/instances/${name}/logs?${new URLSearchParams({ lines: String(lines || 100), ...(level && { level }) })}`
    ),
  health: (name: string, timeout?: number) => {
    const params = new URLSearchParams()
    if (timeout !== undefined) {
      params.set('timeout', String(timeout))
    }
    const query = params.toString()
    return request<InstanceHealthResponse>(
      `/instances/${name}/health${query ? `?${query}` : ''}`,
      { method: 'GET' }
    )
  },
}

export const keyStoreApi = {
  list: () => request<KeyStoreInfo[]>('/security/keystores'),
  get: (name: string) => request<KeyStoreInfo>(`/security/keystores/${name}`),
  create: async (data: {
    name: string
    loaderType: string
    protected?: boolean
    files?: Record<string, File>
  }) => {
    if (data.files && Object.keys(data.files).length > 0) {
      const formData = new FormData()
      formData.append('name', data.name)
      formData.append('loaderType', data.loaderType)
      if (data.protected !== undefined) {
        formData.append('protected', String(data.protected))
      }
      for (const [fieldName, file] of Object.entries(data.files)) {
        formData.append(fieldName, file)
      }
      const response = await fetch(`${API_BASE}/security/keystores`, {
        method: 'POST',
        body: formData,
      })
      if (!response.ok) {
        const text = await response.text()
        throw new Error(text || response.statusText)
      }
      return response.json()
    } else {
      return request<KeyStoreInfo>('/security/keystores', {
        method: 'POST',
        body: JSON.stringify({
          name: data.name,
          loaderType: data.loaderType,
          protected: data.protected,
          params: {},
        }),
      })
    }
  },
  generate: (data: GenerateKeyStoreRequest) =>
    request<KeyStoreInfo>('/security/keystores/generate', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  delete: (name: string) => request<void>(`/security/keystores/${name}`, { method: 'DELETE' }),
  reload: (name: string) => request<void>(`/security/keystores/${name}/reload`, { method: 'POST' }),
}

export const rootCertApi = {
  list: () => request<RootCertInfo[]>('/security/rootcerts'),
  get: (filename: string) => request<RootCertInfo>(`/security/rootcerts/${filename}`),
  add: async (filename: string, certFile: File) => {
    const formData = new FormData()
    formData.append('filename', filename)
    formData.append('cert', certFile)
    const response = await fetch(`${API_BASE}/security/rootcerts`, {
      method: 'POST',
      body: formData,
    })
    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || response.statusText)
    }
    return response.json()
  },
  generate: (data: GenerateRootCARequest) =>
    request<RootCertInfo>('/security/rootcerts/generate', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  delete: (filename: string) => request<void>(`/security/rootcerts/${filename}`, { method: 'DELETE' }),
  reload: () => request<void>('/security/rootcerts/reload', { method: 'POST' }),
}

export const systemApi = {
  info: () => request<SystemInfo>('/system/info'),
  health: () => request<HealthStatus>('/system/health'),
  version: () => request<VersionInfo>('/system/version'),
}

export const configApi = {
  get: () => request<Config>('/config'),
  update: (data: Partial<Config>) =>
    request<Config>('/config', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  reload: () => request<Config>('/config/reload', { method: 'POST' }),
}

export const uiApi = {
  fetchStaticVersion: fetchVersion,
}
