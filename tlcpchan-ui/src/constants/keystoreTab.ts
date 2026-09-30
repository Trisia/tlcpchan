import { CertType } from '@/types'

/**
 * 密钥管理页面的分类 Tab。
 *
 * 页面把 keystore 划分为两类：
 * - PKI：证书类密钥，页面标签为「PKI」，包含 TLCP 的 SM2 双证书与 TLS 的 RSA/ECC 证书
 * - IBC：IBC(SM9) 标识身份密钥，不使用 X.509 证书
 *
 * 取值同时作为 URL query 参数 `tab` 的取值，便于刷新与分享链接保持当前分类。
 */
export const KeystoreTab = {
  /** 证书类密钥（TLCP SM2 双证书 / TLS RSA、ECC 证书） */
  PKI: 'pki',
  /** IBC(SM9) 标识身份密钥 */
  IBC: 'ibc'
} as const

/** 密钥管理页面分类 Tab 的取值类型 */
export type KeystoreTab = typeof KeystoreTab[keyof typeof KeystoreTab]

/** 默认 Tab：URL 未携带 tab 参数或取值非法时使用 */
export const DEFAULT_KEYSTORE_TAB: KeystoreTab = KeystoreTab.PKI

/**
 * 把 URL query 中的 tab 取值规范化为合法的 Tab。
 *
 * query 参数可能缺失、为数组（同名参数多次出现）或用户手工改写的非法值，
 * 这些情况统一回退到 {@link DEFAULT_KEYSTORE_TAB}。
 *
 * @param raw URL query 中的 tab 原始取值，可能是字符串、字符串数组或 undefined
 * @returns 合法的 Tab 取值，非法输入返回默认 Tab
 */
export function normalizeKeyStoreTab(raw: unknown): KeystoreTab {
  const value = Array.isArray(raw) ? raw[0] : raw
  if (value === KeystoreTab.IBC) return KeystoreTab.IBC
  if (value === KeystoreTab.PKI) return KeystoreTab.PKI
  return DEFAULT_KEYSTORE_TAB
}

/**
 * 按 Tab 过滤 keystore 列表。
 *
 * PKI Tab 保留所有证书类密钥（tlcp、tls 以及未来新增的证书类型），
 * IBC Tab 仅保留 type=ibc 的标识身份密钥。
 *
 * @param keystores keystore 元信息列表，元素至少包含 type 字段
 * @param tab 当前选中的分类 Tab
 * @returns 属于该 Tab 的 keystore 列表，输入为空时返回空数组
 */
export function filterKeyStoresByTab<T extends { type: string }>(
  keystores: T[] | undefined | null,
  tab: KeystoreTab
): T[] {
  const list = keystores || []
  if (tab === KeystoreTab.IBC) {
    return list.filter((ks) => ks.type === CertType.IBC)
  }
  return list.filter((ks) => ks.type !== CertType.IBC)
}

/**
 * 构造密钥管理页面的路由目标，使返回后停留在该密钥类型所属的分类 Tab。
 *
 * 生成、导入、替换材料等子页面完成后用它回到列表页，避免 IBC 身份操作后被带回 PKI 分类。
 *
 * @param type keystore 类型（tlcp / tls / ibc）
 * @returns 可用于 router.push 的路由目标，query 中携带 tab 分类
 */
export function keystoreListLocation(type: string): { path: string; query: { tab: KeystoreTab } } {
  return {
    path: '/keystores',
    query: { tab: type === CertType.IBC ? KeystoreTab.IBC : KeystoreTab.PKI }
  }
}