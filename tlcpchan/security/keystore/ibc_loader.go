package keystore

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/security/der"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/sm9"
)

// IBC（SM9）身份材料的参数键名，与 config.KeyStoreConfig.Params 的键一致。
const (
	// IBCParamIdentity 本端标识，如 "server@tlcpchan.local"
	IBCParamIdentity = "identity"
	// IBCParamParams 本端 KGC 公共参数（IBCSysParams）
	IBCParamParams = "params"
	// IBCParamSignKey 签名用户私钥（hid=0x01），PKCS#8
	IBCParamSignKey = "sign-key"
	// IBCParamEncKey 加密用户私钥（hid=0x03），PKCS#8
	IBCParamEncKey = "enc-key"
	// IBCParamKexKey 密钥交换用户私钥（hid=0x02），PKCS#8
	IBCParamKexKey = "kex-key"
)

// SM9 用户私钥的派生用途标识（hid），与 GM/T 0044 定义一致。
const (
	// ibcHidSign 签名私钥用途标识
	ibcHidSign byte = 0x01
)

// IBCFileKeyStore 基于文件的 IBC（SM9）身份密钥存储。
//
// 材料由五个文件构成：标识、本端 KGC 公共参数、签名私钥（hid=0x01）、
// 加密私钥（hid=0x03）、密钥交换私钥（hid=0x02）。其中除标识外均可缺省，
// 缺省意味着本端不具备对应能力（由实例诊断日志提示具体影响）。
//
// 装载结果（身份与元信息）惰性缓存，由 sync.RWMutex 保护。
type IBCFileKeyStore struct {
	identityPath string // 标识文件路径，空表示未提供
	paramsPath   string // 本端 KGC 公共参数文件路径，空表示未提供
	signKeyPath  string // 签名私钥文件路径，空表示未提供
	encKeyPath   string // 加密私钥文件路径，空表示未提供
	kexKeyPath   string // 密钥交换私钥文件路径，空表示未提供

	identity *tlcp.IBCIdentity
	info     *IBCInfo
	loaded   bool
	loadErr  error
	mu       sync.RWMutex
}

// Type 返回 keystore 类型，IBC keystore 固定为 "ibc"。
func (s *IBCFileKeyStore) Type() KeyStoreType {
	return KeyStoreTypeIBC
}

// TLCPCertificate IBC 身份不使用 X.509 证书，恒返回错误。
//
// 返回值：
//   - []*tlcp.Certificate：恒为 nil
//   - error：固定错误，提示应使用 IBC/IBSDH 套件
func (s *IBCFileKeyStore) TLCPCertificate() ([]*tlcp.Certificate, error) {
	return nil, fmt.Errorf("IBC 身份不提供 X.509 证书，请使用 IBC/IBSDH 密码套件")
}

// TLSCertificate IBC 身份不适用于 TLS 协议，恒返回错误。
//
// 返回值：
//   - *tls.Certificate：恒为 nil
//   - error：固定错误
func (s *IBCFileKeyStore) TLSCertificate() (*tls.Certificate, error) {
	return nil, fmt.Errorf("IBC 身份不提供 X.509 证书，不能用于 TLS 协议")
}

// IBCIdentity 返回装载完成的 IBC 身份，首次调用时读取并校验全部材料。
//
// 返回值：
//   - *tlcp.IBCIdentity：装载完成的身份凭据
//   - error：材料读取、解析失败或签名私钥自检失败时返回非 nil
func (s *IBCFileKeyStore) IBCIdentity() (*tlcp.IBCIdentity, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}
	return s.identity, nil
}

// IBCInfo 返回 IBC 身份的只读元信息。
//
// 返回值：
//   - *IBCInfo：元信息；装载失败时返回 nil
func (s *IBCFileKeyStore) IBCInfo() *IBCInfo {
	if err := s.ensureLoaded(); err != nil {
		return nil
	}
	return s.info
}

// ensureLoaded 保证材料只装载一次，并缓存装载结果与错误。
//
// 返回值：
//   - error：首次装载失败时返回该错误，后续调用返回同一错误
func (s *IBCFileKeyStore) ensureLoaded() error {
	s.mu.RLock()
	if s.loaded {
		err := s.loadErr
		s.mu.RUnlock()
		return err
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return s.loadErr
	}

	ident, err := s.load()
	s.identity = ident
	s.loadErr = err
	if err == nil {
		s.info = buildIBCInfo(ident)
	}
	s.loaded = true
	return s.loadErr
}

// load 读取并装载全部 IBC 材料。
//
// 返回值：
//   - *tlcp.IBCIdentity：装载完成的身份凭据
//   - error：文件读取、格式解析或签名私钥自检失败时返回非 nil
func (s *IBCFileKeyStore) load() (*tlcp.IBCIdentity, error) {
	identityData, err := readOptionalFile(s.identityPath)
	if err != nil {
		return nil, fmt.Errorf("读取 IBC 标识文件失败: %w", err)
	}

	var paramsData, signKeyData, encKeyData, kexKeyData []byte
	for _, item := range []struct {
		name string
		path string
		out  *[]byte
	}{
		{"KGC 公共参数", s.paramsPath, &paramsData},
		{"签名私钥", s.signKeyPath, &signKeyData},
		{"加密私钥", s.encKeyPath, &encKeyData},
		{"密钥交换私钥", s.kexKeyPath, &kexKeyData},
	} {
		raw, err := readOptionalFile(item.path)
		if err != nil {
			return nil, fmt.Errorf("读取 IBC %s 文件失败: %w", item.name, err)
		}
		*item.out = raw
	}

	return LoadIBCIdentityFromData(identityData, paramsData, signKeyData, encKeyData, kexKeyData)
}

// LoadIBCIdentityFromData 由内存中的材料字节装载并校验 IBC 身份。
//
// 参数：
//   - identityData: 本端标识，支持裸字节串文本、PEM 或 GM/T 0090 Identifier 的 DER，可为空
//   - paramsData: 本端 KGC 公共参数，支持 PEM / DER / HEX / Base64，可为空
//   - signKeyData: 签名用户私钥（hid=0x01）PKCS#8，支持 PEM / DER / HEX / Base64，可为空
//   - encKeyData: 加密用户私钥（hid=0x03）PKCS#8，编码支持同上，可为空
//   - kexKeyData: 密钥交换用户私钥（hid=0x02）PKCS#8，编码支持同上，可为空
//
// 返回值：
//   - *tlcp.IBCIdentity：装载完成的身份凭据
//   - error：格式解析失败、私钥类型不符或签名私钥自检失败时返回非 nil
//
// 注意事项：
//   - 只做结构解析与签名私钥自检，不阻止"服务端认证场景下客户端不携带私钥"等合法配置
//   - 密钥交换私钥的派生 hid 无法反推校验，正确性由导入来源保证
func LoadIBCIdentityFromData(identityData, paramsData, signKeyData, encKeyData, kexKeyData []byte) (*tlcp.IBCIdentity, error) {
	identity := decodeIBCIdentity(identityData)

	// 公共参数与三把私钥都按 PEM / HEX / Base64 / DER 归一化为 DER，
	// 这样文件内容与内存字节两种来源的行为完全一致
	normalize := func(name string, data []byte) ([]byte, error) {
		if len(data) == 0 {
			return nil, nil
		}
		derData, err := der.Any2DER(data)
		if err != nil {
			return nil, fmt.Errorf("解析 IBC %s 失败: %w", name, err)
		}
		return derData, nil
	}

	paramsDER, err := normalize("KGC 公共参数", paramsData)
	if err != nil {
		return nil, err
	}
	signKeyDER, err := normalize("签名私钥", signKeyData)
	if err != nil {
		return nil, err
	}
	encKeyDER, err := normalize("加密私钥", encKeyData)
	if err != nil {
		return nil, err
	}
	kexKeyDER, err := normalize("密钥交换私钥", kexKeyData)
	if err != nil {
		return nil, err
	}

	ident, err := tlcp.LoadIBCIdentity(identity, paramsDER, signKeyDER, encKeyDER, kexKeyDER)
	if err != nil {
		return nil, err
	}
	if err := verifyIBCSignKey(ident); err != nil {
		return nil, err
	}
	return ident, nil
}

// verifyIBCSignKey 使用公共参数中的签名主公钥对本端标识做一次签名/验签自检。
//
// 参数：
//   - ident: 已装载的身份凭据
//
// 返回值：
//   - error：签名失败（如私钥未携带主公钥信息）或验签不通过时返回错误
//
// 注意事项：
//   - 缺少私钥、公共参数或标识时跳过自检（这些情形由实例诊断日志提示）
//   - 该自检只能覆盖签名私钥，密钥交换私钥（hid=0x02）无法反推校验
func verifyIBCSignKey(ident *tlcp.IBCIdentity) error {
	if ident == nil || ident.SignPrivateKey == nil || ident.Parameters == nil ||
		ident.Parameters.SignMasterPublicKey == nil || len(ident.Identity) == 0 {
		return nil
	}

	digest := sm3.Sum([]byte("tlcpchan-ibc-sign-selfcheck"))
	sig, err := sm9.SignASN1(rand.Reader, ident.SignPrivateKey, digest[:])
	if err != nil {
		return fmt.Errorf("签名私钥自检失败（签名失败，私钥可能缺少主公钥信息）: %w", err)
	}
	if !sm9.VerifyASN1(ident.Parameters.SignMasterPublicKey, ident.Identity, ibcHidSign, digest[:], sig) {
		return fmt.Errorf("签名私钥自检失败：私钥与本端标识或 KGC 公共参数不匹配")
	}
	return nil
}

// buildIBCInfo 由装载完成的身份构造只读元信息。
//
// 参数：
//   - ident: 已装载的身份凭据，可为 nil
//
// 返回值：
//   - *IBCInfo：元信息；ident 为 nil 时返回 nil
func buildIBCInfo(ident *tlcp.IBCIdentity) *IBCInfo {
	if ident == nil {
		return nil
	}
	info := &IBCInfo{
		Identity:          string(ident.Identity),
		HasSignKey:        ident.SignPrivateKey != nil,
		HasEncryptKey:     ident.EncryptPrivateKey != nil,
		HasKeyExchangeKey: ident.KeyExchangePrivateKey != nil,
	}
	if p := ident.Parameters; p != nil {
		info.HasParams = true
		info.DistrictName = p.DistrictName
		info.DistrictSerial = p.DistrictSerial
		info.NotBefore = p.Validity.NotBefore
		info.NotAfter = p.Validity.NotAfter
	}
	return info
}

// decodeIBCIdentity 解析标识输入：优先按 PEM / Identifier DER 解码，否则按裸字节串处理。
//
// 参数：
//   - data: 标识原始文件内容，可为空
//
// 返回值：
//   - []byte: 标识字节串；data 为空时返回 nil
//
// 注意事项：
//   - 不能直接使用 der.Any2DER：裸标识如 "server@tlcpchan.local" 不是合法 DER，
//     但它恰恰是推荐用法（GM/T 0090 Identifier 的 validStart 必填会带来额外配置负担）
func decodeIBCIdentity(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	if block, _ := pem.Decode(data); block != nil && len(block.Bytes) > 0 {
		return block.Bytes
	}
	trimmed := bytes.TrimRight(data, "\r\n")
	if len(trimmed) > 0 && trimmed[0] == 0x30 {
		if derData, err := der.Any2DER(trimmed); err == nil {
			return derData
		}
	}
	return trimmed
}

// readOptionalFile 读取可选文件：路径为空表示未提供，返回 nil 且不报错。
//
// 参数：
//   - path: 文件路径，可为空
//
// 返回值：
//   - []byte: 文件内容；path 为空时返回 nil
//   - error: 路径非空但读取失败时返回错误
func readOptionalFile(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}

// IBCFileLoader IBC（SM9）身份文件加载器。
type IBCFileLoader struct {
	baseDir string
}

// NewIBCFileLoader 创建 IBC 身份文件加载器。
//
// 参数：
//   - baseDir: 相对路径的解析基准目录，空表示按当前工作目录解析
//
// 返回值：
//   - *IBCFileLoader: 加载器实例
func NewIBCFileLoader(baseDir string) *IBCFileLoader {
	return &IBCFileLoader{baseDir: baseDir}
}

// Load 装载 IBC 身份 keystore。
//
// 参数：
//   - loaderType: 加载器类型，固定为 LoaderTypeIBCFile
//   - params: 材料路径参数，键为 IBCParamIdentity / IBCParamParams / IBCParamSignKey /
//     IBCParamEncKey / IBCParamKexKey，值可为绝对路径或相对 baseDir 的路径
//
// 返回值：
//   - KeyStore: IBC keystore 实例
//   - error: 全部材料路径为空、文件读取失败、解析失败或签名私钥自检失败时返回错误
//
// 注意事项：
//   - 装载在创建时立即执行一次，使导入阶段即可发现材料错误；之后结果被缓存
func (l *IBCFileLoader) Load(loaderType LoaderType, params map[string]string) (KeyStore, error) {
	ks := &IBCFileKeyStore{
		identityPath: resolveBasePath(l.baseDir, params[IBCParamIdentity]),
		paramsPath:   resolveBasePath(l.baseDir, params[IBCParamParams]),
		signKeyPath:  resolveBasePath(l.baseDir, params[IBCParamSignKey]),
		encKeyPath:   resolveBasePath(l.baseDir, params[IBCParamEncKey]),
		kexKeyPath:   resolveBasePath(l.baseDir, params[IBCParamKexKey]),
	}

	if ks.identityPath == "" && ks.paramsPath == "" && ks.signKeyPath == "" &&
		ks.encKeyPath == "" && ks.kexKeyPath == "" {
		return nil, fmt.Errorf("IBC 身份材料不能全部为空")
	}

	if _, err := ks.IBCIdentity(); err != nil {
		return nil, err
	}
	return ks, nil
}

// resolveBasePath 将相对路径解析到基准目录下，绝对路径原样返回。
//
// 参数：
//   - baseDir: 基准目录，空表示不拼接
//   - path: 待解析路径，空表示未提供
//
// 返回值：
//   - string: 解析后的路径；path 为空时返回空串
func resolveBasePath(baseDir, path string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) || baseDir == "" {
		return path
	}
	return filepath.Join(baseDir, path)
}
