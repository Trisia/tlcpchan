import { describe, expect, it } from 'vitest'
import { formatDateTime, isZeroTime } from './datetime'

describe('isZeroTime', () => {
  it('识别空值与 Go 时间零值', () => {
    expect(isZeroTime(undefined)).toBe(true)
    expect(isZeroTime(null)).toBe(true)
    expect(isZeroTime('')).toBe(true)
    expect(isZeroTime('0001-01-01T00:00:00Z')).toBe(true)
  })

  it('真实时间返回 false', () => {
    expect(isZeroTime('2026-01-02T03:04:05Z')).toBe(false)
  })
})

describe('formatDateTime', () => {
  it('空值与零值返回替代文案', () => {
    expect(formatDateTime(undefined)).toBe('-')
    expect(formatDateTime('0001-01-01T00:00:00Z')).toBe('-')
    expect(formatDateTime('', '不限')).toBe('不限')
  })

  it('真实时间格式化为本地时间文本', () => {
    expect(formatDateTime('2026-01-02T03:04:05Z')).toBe(
      new Date('2026-01-02T03:04:05Z').toLocaleString('zh-CN')
    )
  })
})