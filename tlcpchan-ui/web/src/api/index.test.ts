import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { systemApi, uiApi } from './index'
import type { VersionInfo, UIVersionInfo } from '@/types'

describe('api/index', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn())
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  describe('systemApi.version', () => {
    it('应该正确获取版本信息', async () => {
      const mockResponse: VersionInfo = {
        version: '1.0.0',
        go_version: 'go1.21.0',
      }

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse),
      } as Response)

      const result = await systemApi.version()
      expect(result).toEqual(mockResponse)
      expect(fetch).toHaveBeenCalledWith('/api/v1/system/version', {
        headers: { 'Content-Type': 'application/json' },
      })
    })

    it('请求失败时应该抛出错误', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        text: () => Promise.resolve('Internal Server Error'),
      } as Response)

      await expect(systemApi.version()).rejects.toThrow('Internal Server Error')
    })
  })

  describe('uiApi.version', () => {
    it('应该正确获取UI版本信息', async () => {
      const mockData: UIVersionInfo = {
        version: '1.0.0',
        go_version: 'go1.21.0',
      }

      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        status: 200,
        json: () =>
          Promise.resolve({
            code: 0,
            message: 'success',
            data: mockData,
          }),
      } as Response)

      const result = await uiApi.version()
      expect(result).toEqual(mockData)
      expect(fetch).toHaveBeenCalledWith('/api/v1/ui/version', {
        headers: { 'Content-Type': 'application/json' },
      })
    })
  })

  describe('uiApi.fetchStaticVersion', () => {
    it('应该正确获取静态版本文件', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: true,
        text: () => Promise.resolve('2.1.0\n'),
      } as Response)

      const result = await uiApi.fetchStaticVersion()
      expect(result).toBe('2.1.0')
      expect(fetch).toHaveBeenCalledWith('/version.txt')
    })

    it('版本文件不存在时应该返回dev', async () => {
      vi.mocked(fetch).mockResolvedValueOnce({
        ok: false,
      } as Response)

      const result = await uiApi.fetchStaticVersion()
      expect(result).toBe('dev')
    })

    it('请求异常时应该返回dev', async () => {
      vi.mocked(fetch).mockRejectedValueOnce(new Error('Network error'))

      const result = await uiApi.fetchStaticVersion()
      expect(result).toBe('dev')
    })
  })
})
