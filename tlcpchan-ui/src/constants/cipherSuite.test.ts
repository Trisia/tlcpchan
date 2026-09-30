import { describe, expect, it } from 'vitest'
import {
  TLCP_CIPHER_SUITES,
  TLS_CIPHER_SUITES,
  cipherSuiteDisabledReason,
  isClientAuthAllowedForECDHE,
  isIBCSuite
} from './cipherSuite'

describe('TLCP_CIPHER_SUITES', () => {
  it('包含证书套件与 4 个默认关闭的 IBC/IBSDH 套件', () => {
    expect(TLCP_CIPHER_SUITES).toEqual([
      'ECC_SM4_CBC_SM3',
      'ECC_SM4_GCM_SM3',
      'ECDHE_SM4_CBC_SM3',
      'ECDHE_SM4_GCM_SM3',
      'IBC_SM4_GCM_SM3',
      'IBC_SM4_CBC_SM3',
      'IBSDH_SM4_GCM_SM3',
      'IBSDH_SM4_CBC_SM3'
    ])
  })

  it('IBSDH 套件以 IBC 前缀之外的名称出现，判定不能只依赖 IBC_ 前缀', () => {
    expect(TLCP_CIPHER_SUITES.filter((suite) => suite.startsWith('IBSDH_'))).toEqual([
      'IBSDH_SM4_GCM_SM3',
      'IBSDH_SM4_CBC_SM3'
    ])
  })
})

describe('TLS_CIPHER_SUITES', () => {
  it('不包含任何 IBC/IBSDH 套件', () => {
    expect(TLS_CIPHER_SUITES.some((suite) => isIBCSuite(suite))).toBe(false)
  })
})

describe('isIBCSuite', () => {
  it('4 个 IBC/IBSDH 套件均判定为 IBC 套件', () => {
    expect(isIBCSuite('IBC_SM4_GCM_SM3')).toBe(true)
    expect(isIBCSuite('IBC_SM4_CBC_SM3')).toBe(true)
    expect(isIBCSuite('IBSDH_SM4_GCM_SM3')).toBe(true)
    expect(isIBCSuite('IBSDH_SM4_CBC_SM3')).toBe(true)
  })

  it('证书套件与未知取值均判定为非 IBC 套件', () => {
    expect(isIBCSuite('ECC_SM4_GCM_SM3')).toBe(false)
    expect(isIBCSuite('ECDHE_SM4_CBC_SM3')).toBe(false)
    expect(isIBCSuite('TLS_AES_128_GCM_SHA256')).toBe(false)
    expect(isIBCSuite('')).toBe(false)
  })
})

describe('cipherSuiteDisabledReason', () => {
  it('未配置 IBC 身份时 IBC/IBSDH 套件给出提示', () => {
    expect(cipherSuiteDisabledReason('IBC_SM4_GCM_SM3', false)).toBe('需先配置 IBC 身份 keystore')
    expect(cipherSuiteDisabledReason('IBSDH_SM4_CBC_SM3', false)).toBe('需先配置 IBC 身份 keystore')
  })

  it('已配置 IBC 身份时 IBC 套件可选', () => {
    expect(cipherSuiteDisabledReason('IBC_SM4_GCM_SM3', true)).toBe('')
    expect(cipherSuiteDisabledReason('IBSDH_SM4_GCM_SM3', true)).toBe('')
  })

  it('证书套件与 IBC 身份无关，始终可选', () => {
    expect(cipherSuiteDisabledReason('ECC_SM4_GCM_SM3', false)).toBe('')
    expect(cipherSuiteDisabledReason('ECDHE_SM4_CBC_SM3', false)).toBe('')
  })
})

describe('isClientAuthAllowedForECDHE', () => {
  it('要求客户端出示证书的两种认证类型满足 ECDHE 要求', () => {
    expect(isClientAuthAllowedForECDHE('require-any-client-cert')).toBe(true)
    expect(isClientAuthAllowedForECDHE('require-and-verify-client-cert')).toBe(true)
  })

  it('不强制客户端出示证书的认证类型不满足 ECDHE 要求', () => {
    expect(isClientAuthAllowedForECDHE('no-client-cert')).toBe(false)
    expect(isClientAuthAllowedForECDHE('request-client-cert')).toBe(false)
    expect(isClientAuthAllowedForECDHE('verify-client-cert-if-given')).toBe(false)
  })

  it('空值与未知取值按后端默认值 no-client-cert 处理，不满足 ECDHE 要求', () => {
    expect(isClientAuthAllowedForECDHE('')).toBe(false)
    expect(isClientAuthAllowedForECDHE(undefined)).toBe(false)
    expect(isClientAuthAllowedForECDHE('unknown-auth')).toBe(false)
  })
})