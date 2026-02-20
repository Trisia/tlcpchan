import axios from 'axios'
import API_CONFIG from '@/config/api'

const http = axios.create({
  baseURL: API_CONFIG.baseURL,
  timeout: API_CONFIG.timeout,
  headers: API_CONFIG.headers,
})

export default http

// 获取实例列表
export function getInstances() {
  return http.get('/instances').then((data: any) => data.instances || [])
}

// 获取单个实例
export function getInstance(name: string) {
  return http.get(`/instances/${name}`)
}

// 创建实例
export function createInstance(data: any) {
  return http.post('/instances', data)
}

// 更新实例
export function updateInstance(name: string, data: any) {
  return http.put(`/instances/${name}`, data)
}

// 删除实例
export function deleteInstance(name: string) {
  return http.delete(`/instances/${name}`)
}

// 启动实例
export function startInstance(name: string) {
  return http.post(`/instances/${name}/start`)
}

// 停止实例
export function stopInstance(name: string) {
  return http.post(`/instances/${name}/stop`)
}

// 重载实例
export function reloadInstance(name: string) {
  return http.post(`/instances/${name}/reload`)
}

// 获取实例统计
export function getInstanceStats(name: string) {
  return http.get(`/instances/${name}/stats`)
}

// 获取实例日志
export function getInstanceLogs(name: string, lines?: number, level?: string) {
  const params: any = {}
  if (lines) params.lines = lines
  if (level) params.level = level
  return http.get(`/instances/${name}/logs`, { params })
}

// 获取实例健康状态
export function getInstanceHealth(name: string) {
  return http.get(`/instances/${name}/health`)
}

export default http