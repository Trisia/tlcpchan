/**
 * 客户端认证类型（client-auth-type）的中文文案与判定工具。
 *
 * 英文取值与后端 tlcpchan/config/config.go 中 ValidClientAuthValues 的枚举保持一致，
 * 前端统一以“中文（英文）”的形式呈现，避免用户直接面对裸英文枚举。
 */

/** 客户端认证类型英文取值 → 中文文案 */
const CLIENT_AUTH_LABELS: Record<string, string> = {
  'no-client-cert': '不要求证书',
  'request-client-cert': '请求证书',
  'require-any-client-cert': '要求证书',
  'verify-client-cert-if-given': '提供则验证',
  'require-and-verify-client-cert': '要求并验证',
}

/**
 * 将客户端认证类型格式化为“中文（英文）”展示文本。
 *
 * 参数:
 *   - value: 认证类型英文取值，为空时按后端默认值 "no-client-cert" 处理
 *
 * 返回:
 *   - 形如“不要求证书（no-client-cert）”的字符串；取值未知时原样返回该取值
 */
export function formatClientAuth(value?: string): string {
  const key = value || 'no-client-cert'
  const label = CLIENT_AUTH_LABELS[key]
  return label ? `${label}（${key}）` : key
}

/**
 * 判断客户端认证类型是否为单向认证。
 *
 * 单向认证指对端不强制本端出示客户端证书，即本端 keystore 可选：
 * "no-client-cert" 完全不请求客户端证书，"request-client-cert" 仅请求但不强制。
 *
 * 参数:
 *   - value: 认证类型英文取值，为空时按后端默认值 "no-client-cert" 处理
 *
 * 返回:
 *   - true 表示单向认证；false 表示需要（或可能需要）本端提供客户端证书
 */
export function isOneWayClientAuth(value?: string): boolean {
  const key = value || 'no-client-cert'
  return key === 'no-client-cert' || key === 'request-client-cert'
}