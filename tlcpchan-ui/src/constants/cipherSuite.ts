/**
 * TLCP/TLS 密码套件常量与 IBC（SM9）套件判定工具。
 *
 * 套件名称与后端 tlcpchan/config/config.go 中的 TLCPCipherSuiteNames / TLSCipherSuiteNames
 * 保持一致；IBC/IBSDH 套件默认关闭，必须显式勾选且实例配置了 IBC 身份后才会参与协商。
 *
 * ECDHE 套件（ECDHE_SM4_*）要求服务端对客户端做身份认证，客户端认证类型必须是
 * “要求证书”或“要求并验证”，该判定由 isClientAuthAllowedForECDHE 提供。
 */

/** TLCP 可选密码套件（前 4 个为证书套件，后 4 个为默认关闭的 IBC/IBSDH 套件） */
export const TLCP_CIPHER_SUITES = [
  'ECC_SM4_CBC_SM3',
  'ECC_SM4_GCM_SM3',
  'ECDHE_SM4_CBC_SM3',
  'ECDHE_SM4_GCM_SM3',
  'IBC_SM4_GCM_SM3',
  'IBC_SM4_CBC_SM3',
  'IBSDH_SM4_GCM_SM3',
  'IBSDH_SM4_CBC_SM3'
]

/** TLS 可选密码套件 */
export const TLS_CIPHER_SUITES = [
  'TLS_RSA_WITH_AES_128_GCM_SHA256',
  'TLS_RSA_WITH_AES_256_GCM_SHA384',
  'TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256',
  'TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384',
  'TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256',
  'TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384',
  'TLS_AES_128_GCM_SHA256',
  'TLS_AES_256_GCM_SHA384',
  'TLS_CHACHA20_POLY1305_SHA256'
]

/** IBC/IBSDH 套件集合，用于统一判定是否需要 IBC 身份 */
const IBC_CIPHER_SUITES = new Set([
  'IBC_SM4_GCM_SM3',
  'IBC_SM4_CBC_SM3',
  'IBSDH_SM4_GCM_SM3',
  'IBSDH_SM4_CBC_SM3'
])

/**
 * 判断套件是否为 IBC/IBSDH 套件。
 *
 * 参数:
 *   - suite: 套件名称
 *
 * 返回:
 *   - 是 IBC/IBSDH 套件时返回 true
 */
export function isIBCSuite(suite: string): boolean {
  return IBC_CIPHER_SUITES.has(suite)
}

/**
 * 满足 ECDHE 套件要求的客户端认证类型集合。
 *
 * ECDHE 套件必须对客户端做身份认证，因此服务端必须"要求"客户端出示证书：
 * - require-any-client-cert：要求证书（不校验证书链）
 * - require-and-verify-client-cert：要求并验证
 * 不包含 no-client-cert（不要求）、request-client-cert（仅请求，不强制）、
 * verify-client-cert-if-given（提供才验证，不强制出示）。
 */
const ECDHE_CLIENT_AUTH_TYPES = new Set([
  'require-any-client-cert',
  'require-and-verify-client-cert'
])

/**
 * 判断客户端认证类型是否满足 ECDHE 套件对客户端身份认证的要求。
 *
 * 参数:
 *   - clientAuthType: 客户端认证类型英文取值，为空时按后端默认值 "no-client-cert" 处理
 *
 * 返回:
 *   - 满足（要求客户端出示证书）时返回 true
 */
export function isClientAuthAllowedForECDHE(clientAuthType?: string): boolean {
  return ECDHE_CLIENT_AUTH_TYPES.has(clientAuthType || '')
}

/**
 * 计算套件不可选的原因，用于表单中置灰复选框并给出提示。
 *
 * 参数:
 *   - suite: 套件名称
 *   - hasIBCKeystore: 实例是否已选择 IBC 身份密钥库
 *
 * 返回:
 *   - 不可选时返回提示文案；可选时返回空字符串
 */
export function cipherSuiteDisabledReason(suite: string, hasIBCKeystore: boolean): string {
  if (isIBCSuite(suite) && !hasIBCKeystore) return '需先配置 IBC 身份 keystore'
  return ''
}