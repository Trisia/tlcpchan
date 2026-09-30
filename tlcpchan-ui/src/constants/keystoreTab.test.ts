import { describe, expect, it } from 'vitest'
import {
  DEFAULT_KEYSTORE_TAB,
  KeystoreTab,
  filterKeyStoresByTab,
  keystoreListLocation,
  normalizeKeyStoreTab
} from './keystoreTab'

describe('normalizeKeyStoreTab', () => {
  it('接受两个合法 Tab 取值', () => {
    expect(normalizeKeyStoreTab('pki')).toBe(KeystoreTab.PKI)
    expect(normalizeKeyStoreTab('ibc')).toBe(KeystoreTab.IBC)
  })

  it('缺失或非法取值回退到默认 Tab', () => {
    expect(normalizeKeyStoreTab(undefined)).toBe(DEFAULT_KEYSTORE_TAB)
    expect(normalizeKeyStoreTab(null)).toBe(DEFAULT_KEYSTORE_TAB)
    expect(normalizeKeyStoreTab('')).toBe(DEFAULT_KEYSTORE_TAB)
    expect(normalizeKeyStoreTab('sm2')).toBe(DEFAULT_KEYSTORE_TAB)
    // 旧的 cert 取值不再识别，按非法值处理
    expect(normalizeKeyStoreTab('cert')).toBe(DEFAULT_KEYSTORE_TAB)
  })

  it('同名 query 参数出现多次时取第一个', () => {
    expect(normalizeKeyStoreTab(['ibc', 'pki'])).toBe(KeystoreTab.IBC)
    // 第一个取值非法时仍回退默认，不继续向后查找
    expect(normalizeKeyStoreTab(['unknown', 'ibc'])).toBe(DEFAULT_KEYSTORE_TAB)
  })
})

describe('filterKeyStoresByTab', () => {
  const keystores = [
    { name: 'tlcp-server', type: 'tlcp' },
    { name: 'tls-server', type: 'tls' },
    { name: 'ibc-server', type: 'ibc' }
  ]

  it('PKI Tab 保留全部证书类密钥', () => {
    expect(filterKeyStoresByTab(keystores, KeystoreTab.PKI).map((ks) => ks.name)).toEqual([
      'tlcp-server',
      'tls-server'
    ])
  })

  it('IBC Tab 仅保留标识身份密钥', () => {
    expect(filterKeyStoresByTab(keystores, KeystoreTab.IBC).map((ks) => ks.name)).toEqual([
      'ibc-server'
    ])
  })

  it('列表为空或未加载时返回空数组', () => {
    expect(filterKeyStoresByTab([], KeystoreTab.PKI)).toEqual([])
    expect(filterKeyStoresByTab(undefined, KeystoreTab.IBC)).toEqual([])
    expect(filterKeyStoresByTab(null, KeystoreTab.IBC)).toEqual([])
  })
})

describe('keystoreListLocation', () => {
  it('IBC 类型返回 IBC 分类', () => {
    expect(keystoreListLocation('ibc')).toEqual({
      path: '/keystores',
      query: { tab: KeystoreTab.IBC }
    })
  })

  it('证书类型返回 PKI 分类', () => {
    expect(keystoreListLocation('tlcp').query.tab).toBe(KeystoreTab.PKI)
    expect(keystoreListLocation('tls').query.tab).toBe(KeystoreTab.PKI)
  })
})