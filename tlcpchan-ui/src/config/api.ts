/**
 * API配置
 * 
 * 开发环境: /api
 * 生产环境: ../api
 */

export const API_CONFIG = {
  baseURL: import.meta.env.VITE_API_BASE || '/api',
  timeout: parseInt(import.meta.env.VITE_API_TIMEOUT || '10000'),
  headers: {
    'Content-Type': 'application/json',
  },
}

export default API_CONFIG