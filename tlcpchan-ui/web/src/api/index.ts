import type { Instance, Certificate, SystemInfo, HealthInfo, VersionInfo, UIVersionInfo } from '@/types'

const API_BASE = '/api/v1'

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
  stats: (name: string, period?: string) =>
    request<{ connections_total: number; connections_active: number; bytes_received: number; bytes_sent: number; latency_avg_ms: number }>(
      `/instances/${name}/stats${period ? `?period=${period}` : ''}`
    ),
  logs: (name: string, lines?: number, level?: string) =>
    request<{ logs: Array<{ time: string; level: string; message: string }> }>(
      `/instances/${name}/logs?${new URLSearchParams({ lines: String(lines || 100), ...(level && { level }) })}`
    ),
}

export const certificateApi = {
  list: () => request<{ certificates: Certificate[]; total: number }>('/certificates'),
  get: (name: string) => request<Certificate>(`/certificates/${name}`),
  reload: () => request<{ reloaded: boolean; updated: string[] }>('/certificates/reload', { method: 'POST' }),
  generate: (data: {
    type: 'tlcp' | 'tls'
    name: string
    common_name: string
    dns_names?: string[]
    ip_addresses?: string[]
    days?: number
    ca_name?: string
  }) => request<{ name: string; not_before: string; not_after: string }>('/certificates/generate', { method: 'POST', body: JSON.stringify(data) }),
  delete: (name: string) => request<void>(`/certificates/${name}`, { method: 'DELETE' }),
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
