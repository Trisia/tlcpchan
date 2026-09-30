import { describe, expect, it } from 'vitest'
import { formatClientAuth, isOneWayClientAuth } from './clientAuth'

describe('formatClientAuth', () => {
  it('全部五种认证类型均以“中文（英文）”呈现', () => {
    expect(formatClientAuth('no-client-cert')).toBe('不要求证书（no-client-cert）')
    expect(formatClientAuth('request-client-cert')).toBe('请求证书（request-client-cert）')
    expect(formatClientAuth('require-any-client-cert')).toBe('要求证书（require-any-client-cert）')
    expect(formatClientAuth('verify-client-cert-if-given')).toBe('提供则验证（verify-client-cert-if-given）')
    expect(formatClientAuth('require-and-verify-client-cert')).toBe(
      '要求并验证（require-and-verify-client-cert）'
    )
  })

  it('空值按默认值 no-client-cert 呈现', () => {
    expect(formatClientAuth('')).toBe('不要求证书（no-client-cert）')
    expect(formatClientAuth(undefined)).toBe('不要求证书（no-client-cert）')
  })

  it('未知取值原样返回，不伪造中文文案', () => {
    expect(formatClientAuth('unknown-auth')).toBe('unknown-auth')
  })
})

describe('isOneWayClientAuth', () => {
  it('不要求证书与请求证书视为单向认证', () => {
    expect(isOneWayClientAuth('no-client-cert')).toBe(true)
    expect(isOneWayClientAuth('request-client-cert')).toBe(true)
  })

  it('空值按默认值 no-client-cert，视为单向认证', () => {
    expect(isOneWayClientAuth('')).toBe(true)
    expect(isOneWayClientAuth(undefined)).toBe(true)
  })

  it('要求类与验证类认证类型不属于单向认证', () => {
    expect(isOneWayClientAuth('require-any-client-cert')).toBe(false)
    expect(isOneWayClientAuth('verify-client-cert-if-given')).toBe(false)
    expect(isOneWayClientAuth('require-and-verify-client-cert')).toBe(false)
  })
})