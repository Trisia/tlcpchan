# 国家电子认证根CA运营证书

本目录包含从国家电子认证根CA网站下载的状态为"正常"的证书。

## 证书来源

- **网站地址**: http://www.rootca.gov.cn/runningCa.jsp
- **下载时间**: 2026年2月15日
- **证书数量**: 86个
- **证书状态**: 正常（normal）

## 证书格式

所有证书均为标准PEM格式，包含以下头部和尾部：
```
-----BEGIN CERTIFICATE-----
... Base64编码的证书内容 ...
-----END CERTIFICATE-----
```

## 使用说明

这些证书可用于TLCP/TLS协议中的信任链验证，支持SM2、RSA等算法。

## 注意事项

请定期访问 http://www.rootca.gov.cn/runningCa.jsp 更新证书，以确保使用最新的有效证书。
