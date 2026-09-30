/**
 * 后端时间的展示工具。
 *
 * 后端 JSON 中的时间统一为 RFC 3339 字符串。IBC 元信息中「未提供公共参数」时，
 * Go 的 time.Time 会序列化为零值 `0001-01-01T00:00:00Z`（而非空串或 null），
 * 因此展示前必须先区分零值与真实时间，避免界面出现 0001 年。
 */

/**
 * 判断后端返回的时间是否为零值（语义为“未提供/不限”）
 * @param dateStr RFC 3339 时间字符串，可能为空
 * @returns 空值、空串或以 0001-01-01 开头的 Go 零值返回 true
 */
export function isZeroTime(dateStr: string | undefined | null): boolean {
  return !dateStr || dateStr.startsWith('0001-01-01')
}

/**
 * 把 RFC 3339 时间字符串格式化为本地时间文本
 * @param dateStr RFC 3339 时间字符串，可能为空
 * @param fallback 空值或零值时返回的替代文案，默认 '-'
 * @returns 本地化时间文本（如 2026/1/1 12:00:00），或 fallback
 */
export function formatDateTime(dateStr: string | undefined | null, fallback = '-'): string {
  if (isZeroTime(dateStr)) return fallback
  return new Date(dateStr as string).toLocaleString('zh-CN')
}