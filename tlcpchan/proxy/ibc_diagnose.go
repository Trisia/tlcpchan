package proxy

import (
	"fmt"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security"
)

// 诊断级别常量，用于区分"握手必然失败"与"可能受影响"。
const (
	// ibcDiagLevelError 表示该配置下的 IBC/IBSDH 握手必然失败
	ibcDiagLevelError = "error"
	// ibcDiagLevelWarn 表示该配置下握手可能受影响，或仅在部分场景下失败
	ibcDiagLevelWarn = "warn"
)

// ibcDiagnosis 单条 IBC 能力诊断结论。
type ibcDiagnosis struct {
	// Level 诊断级别，取值为 ibcDiagLevelError 或 ibcDiagLevelWarn
	Level string
	// Message 诊断说明，包含实例名称、角色与具体缺失材料
	Message string
}

// diagnoseIBCSuites 诊断实例的 IBC 套件配置与本端 IBC 能力是否匹配，仅产出诊断结论不阻断启动。
//
// 参数：
//   - instanceName: 实例名称，用于诊断信息定位
//   - isServer: 本端是否为服务端，决定必需材料集合
//   - suites: 已解析的 TLCP 密码套件数值列表
//   - ibcKS: 已加载的 IBC keystore，nil 表示未配置 ibc-keystore
//   - ident: 已装载的 IBC 身份，nil 表示身份不可用（未配置或装载失败）
//   - poolSize: 全局 IBC 信任池中的 KGC 公共参数条目数，0 表示无可信 KGC
//
// 返回值：
//   - []ibcDiagnosis: 诊断结论列表；配置正常或未配置 IBC/IBSDH 套件时为空
//
// 注意事项：
//   - 未配置任何 IBC/IBSDH 套件时直接返回空列表：按方案 D5，"配了 IBC 身份但未勾选 IBC 套件"属正常配置
//   - 本函数不阻断实例启动，库会自动跳过本端不具备能力的套件
func diagnoseIBCSuites(instanceName string, isServer bool, suites []uint16,
	ibcKS security.KeyStore, ident *tlcp.IBCIdentity, poolSize int) []ibcDiagnosis {

	var ibcSuiteNames []string
	hasIBC := false
	hasIBSDH := false
	for _, suite := range suites {
		if !config.IsTLCPIBCSuite(suite) {
			continue
		}
		ibcSuiteNames = append(ibcSuiteNames, tlcp.CipherSuiteName(suite))
		if config.IsTLCPIBSDHSuite(suite) {
			hasIBSDH = true
		} else {
			hasIBC = true
		}
	}
	if len(ibcSuiteNames) == 0 {
		return nil
	}

	role := "客户端"
	if isServer {
		role = "服务端"
	}

	var diagnoses []ibcDiagnosis

	// 场景 1：配套件但未配 IBC 身份
	if ibcKS == nil || ibcKS.Type() != security.KeyStoreTypeIBC {
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
			"实例 %s (%s) 配置了 IBC/IBSDH 密码套件 %v，但未配置可用的 IBC 身份 keystore（tlcp.ibc-keystore），这些套件不会参与协商",
			instanceName, role, ibcSuiteNames)})
		return diagnoses
	}

	// 场景 2：IBC 身份装载失败（详情已在适配器中记录）
	if ident == nil {
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
			"实例 %s (%s) 的 IBC 身份装载失败，IBC/IBSDH 套件 %v 不会参与协商",
			instanceName, role, ibcSuiteNames)})
		return diagnoses
	}

	// 场景 3：标识为空
	switch {
	case len(ident.Identity) == 0 && hasIBSDH:
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelWarn, fmt.Sprintf(
			"实例 %s (%s) 的 IBC 标识为空，IBSDH 握手时对端将报 identity_need(205)", instanceName, role)})
	case len(ident.Identity) == 0:
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelWarn, fmt.Sprintf(
			"实例 %s (%s) 的 IBC 标识为空，对端可能无法确定本端标识", instanceName, role)})
	}

	// 场景 4：缺少本端 KGC 公共参数
	switch {
	case ident.Parameters != nil:
	case isServer || ident.SignPrivateKey != nil:
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
			"实例 %s (%s) 缺少本端 KGC 公共参数，握手将报 bad_ibcparam(203)，请先导入本端公共参数", instanceName, role)})
	default:
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelWarn, fmt.Sprintf(
			"实例 %s (客户端) 未提供本端 KGC 公共参数，仅服务端单向认证场景可省略", instanceName)})
	}

	// 场景 5：缺少签名私钥
	if ident.SignPrivateKey == nil {
		if isServer {
			diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
				"实例 %s (服务端) 缺少签名私钥(hid=0x01)，无法对 signed_params 签名，IBC/IBSDH 握手将失败", instanceName)})
		} else {
			diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelWarn, fmt.Sprintf(
				"实例 %s (客户端) 缺少签名私钥(hid=0x01)，若服务端要求客户端认证将失败", instanceName)})
		}
	}

	// 场景 6：IBC 套件缺少加密私钥（仅服务端需要解密预主密钥）
	if hasIBC && isServer && ident.EncryptPrivateKey == nil {
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
			"实例 %s (服务端) 使用了 IBC_SM4_* 套件但缺少加密私钥(hid=0x03)，无法解密预主密钥，握手将失败", instanceName)})
	}

	// 场景 7：IBSDH 套件缺少密钥交换私钥
	if hasIBSDH && ident.KeyExchangePrivateKey == nil {
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
			"实例 %s (%s) 使用了 IBSDH_SM4_* 套件但缺少密钥交换私钥(hid=0x02)，这些套件不会参与协商", instanceName, role)})
	}

	// 场景 8：全局信任池为空
	if poolSize == 0 {
		diagnoses = append(diagnoses, ibcDiagnosis{ibcDiagLevelError, fmt.Sprintf(
			"实例 %s (%s) 的全局 IBC 信任池为空，无法校验对端 KGC 公共参数，IBC/IBSDH 握手将失败；请先在「IBC 信任池」中添加 KGC 公共参数",
			instanceName, role)})
	}

	return diagnoses
}

// logIBCDiagnoses 按诊断级别输出 IBC 能力诊断日志。
//
// 参数：
//   - diagnoses: diagnoseIBCSuites 产出的诊断结论列表
//
// 注意事项：
//   - error 级别表示握手必然失败，warn 级别表示可能受影响
func logIBCDiagnoses(diagnoses []ibcDiagnosis) {
	for _, diagnosis := range diagnoses {
		if diagnosis.Level == ibcDiagLevelError {
			logger.Error("%s", diagnosis.Message)
		} else {
			logger.Warn("%s", diagnosis.Message)
		}
	}
}
