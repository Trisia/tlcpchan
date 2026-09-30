package ibcparams

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/security/der"
	"github.com/emmansun/gmsm/sm3"
)

// ParamExtensions 信任池目录中参与扫描的文件扩展名白名单（含点，小写）
var ParamExtensions = []string{".pem", ".der", ".ibcparams"}

// Manager IBC 信任池管理器。
//
// 负责扫描信任池目录、增删查 KGC 公共参数、维护用于握手的 tlcp.IBCPool。
// 管理方式与根证书管理器一致：目录扫描 + 扩展名白名单 + 无效文件跳过并记日志。
//
// 安全语义（默认拒绝）：池中不存在对端 KGC 时 IBC 握手直接失败，
// 与"没有根证书就验不过证书"对等。
type Manager struct {
	baseDir string               // 信任池目录，默认 <workDir>/ibcparams
	params  map[string]*IBCParam // 文件名 -> 参数元信息
	pool    *tlcp.IBCPool        // 全量信任池，Reload 时重建
	mu      sync.RWMutex
}

// NewManager 创建 IBC 信任池管理器。
//
// 参数：
//   - baseDir: 信任池目录路径，空表示不扫描任何目录
//
// 返回值：
//   - *Manager: 管理器实例，内部信任池为空池
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir: baseDir,
		params:  make(map[string]*IBCParam),
		pool:    tlcp.NewIBCPool(),
	}
}

// Initialize 初始化管理器，扫描并加载目录中的全部 KGC 公共参数。
//
// 返回值：
//   - error: 目录读取失败时返回错误；目录不存在视为空池
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadAll()
}

// Add 添加 KGC 公共参数（保存到目录并重新加载）。
//
// 参数：
//   - filename: 保存的文件名，必须为不含路径分隔符的纯文件名
//   - data: 公共参数数据，支持 PEM / DER / HEX / Base64
//
// 返回值：
//   - *IBCParam: 添加成功的条目元信息
//   - error: 文件名非法、数据解析失败或与已有条目属于同一 KGC 时返回错误
func (m *Manager) Add(filename string, data []byte) (*IBCParam, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	if !isParamExtension(strings.ToLower(filepath.Ext(filename))) {
		return nil, fmt.Errorf("文件名扩展名必须为 %s 之一", strings.Join(ParamExtensions, " / "))
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("公共参数数据不能为空")
	}

	param, err := parseParam(data, filename)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, existing := range m.params {
		if name == filename {
			continue
		}
		if existing.DistrictName == param.DistrictName && existing.DistrictSerial == param.DistrictSerial {
			return nil, fmt.Errorf("同一 KGC (%s#%d) 已存在: %s", param.DistrictName, param.DistrictSerial, name)
		}
	}

	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(filepath.Join(m.baseDir, filename), data, 0600); err != nil {
		return nil, fmt.Errorf("写入公共参数失败: %w", err)
	}

	if err := m.loadAll(); err != nil {
		return nil, err
	}
	return m.params[filename], nil
}

// Delete 删除 KGC 公共参数。
//
// 参数：
//   - filename: 待删除的文件名
//
// 返回值：
//   - error: 删除失败或重新加载失败时返回错误
func (m *Manager) Delete(filename string) error {
	if err := validateFilename(filename); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	paramPath := filepath.Join(m.baseDir, filename)
	if err := os.Remove(paramPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除公共参数失败: %w", err)
	}

	return m.loadAll()
}

// Get 获取指定文件名的条目元信息。
//
// 参数：
//   - filename: 文件名
//
// 返回值：
//   - *IBCParam: 条目元信息
//   - error: 条目不存在时返回错误
func (m *Manager) Get(filename string) (*IBCParam, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	param, exists := m.params[filename]
	if !exists {
		return nil, fmt.Errorf("IBC 公共参数 %s 不存在", filename)
	}
	return param, nil
}

// List 列出全部条目元信息。
//
// 返回值：
//   - []*IBCParam: 条目元信息列表，池为空时返回长度为 0 的切片
func (m *Manager) List() []*IBCParam {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*IBCParam, 0, len(m.params))
	for _, param := range m.params {
		result = append(result, param)
	}
	return result
}

// GetPool 获取当前的全量信任池。
//
// 返回值：
//   - *tlcp.IBCPool: 信任池；Reload 后会替换为新对象，因此修改信任池后需重载相关实例才生效
//
// 注意事项：
//   - 调用方不得修改返回的池内容，如需更新请调用 Reload
func (m *Manager) GetPool() *tlcp.IBCPool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.pool
}

// Reload 重新扫描目录并重建信任池。
//
// 返回值：
//   - error: 目录读取失败时返回错误
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadAll()
}

// ReadFile 读取信任池中的原始文件内容。
//
// 参数：
//   - filename: 文件名
//
// 返回值：
//   - []byte: 文件原始内容
//   - error: 文件名非法或读取失败时返回错误
func (m *Manager) ReadFile(filename string) ([]byte, error) {
	if err := validateFilename(filename); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(m.baseDir, filename))
}

// loadAll 重新扫描目录、解析全部有效文件并重建信任池。
//
// 返回值：
//   - error: 目录读取失败时返回错误
//
// 注意事项：
//   - 无效文件、重复 KGC 与解析失败的文件会被跳过并记录日志，不影响其余条目
//   - 子目录不参与扫描（信任池只接受顶层文件）
func (m *Manager) loadAll() error {
	m.params = make(map[string]*IBCParam)
	m.pool = tlcp.NewIBCPool()

	if m.baseDir == "" {
		return nil
	}

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取目录失败: %w", err)
	}

	seenKGC := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !isParamExtension(strings.ToLower(filepath.Ext(filename))) {
			continue
		}

		data, err := os.ReadFile(filepath.Join(m.baseDir, filename))
		if err != nil {
			logger.Warn("读取 IBC 公共参数 %s 失败，已跳过: %v", filename, err)
			continue
		}

		param, err := parseParam(data, filename)
		if err != nil {
			logger.Warn("解析 IBC 公共参数 %s 失败，已跳过: %v", filename, err)
			continue
		}

		key := kgcKey(param.DistrictName, param.DistrictSerial)
		if prev, ok := seenKGC[key]; ok {
			logger.Warn("IBC 公共参数 %s 与 %s 属于同一 KGC (%s#%d)，已跳过",
				filename, prev, param.DistrictName, param.DistrictSerial)
			continue
		}
		seenKGC[key] = filename

		if err := m.pool.AddParams(param.Params); err != nil {
			logger.Warn("将 IBC 公共参数 %s 加入信任池失败，已跳过: %v", filename, err)
			continue
		}
		m.params[filename] = param
	}

	return nil
}

// parseParam 解析公共参数数据并构造条目元信息。
//
// 参数：
//   - data: 公共参数数据，支持 PEM / DER / HEX / Base64
//   - filename: 文件名，用于填充 IBCParam.Filename
//
// 返回值：
//   - *IBCParam: 解析成功后的条目元信息
//   - error: 数据格式无法识别或不是合法的 IBCSysParams 时返回错误
func parseParam(data []byte, filename string) (*IBCParam, error) {
	derData, err := der.Any2DER(data)
	if err != nil {
		return nil, fmt.Errorf("解析公共参数格式失败: %w", err)
	}

	params, err := tlcp.ParseIBCSysParams(derData)
	if err != nil {
		return nil, fmt.Errorf("解析 IBCSysParams 失败: %w", err)
	}

	param := &IBCParam{
		Filename:       filename,
		DistrictName:   params.DistrictName,
		DistrictSerial: params.DistrictSerial,
		NotBefore:      params.Validity.NotBefore,
		NotAfter:       params.Validity.NotAfter,
		Params:         params,
	}
	if params.IssuerID != nil {
		param.IssuerIdentity = string(params.IssuerID.IdentityData)
	}
	if params.SignMasterPublicKey != nil {
		param.SignKeyFingerprint = masterPublicKeyFingerprint(params.SignMasterPublicKey)
	}
	if params.EncryptMasterPublicKey != nil {
		param.EncKeyFingerprint = masterPublicKeyFingerprint(params.EncryptMasterPublicKey)
	}
	return param, nil
}

// masterPublicKeyFingerprint 计算 SM9 主公钥的 SM3 指纹。
//
// 参数：
//   - pub: 主公钥，支持 *sm9.SignMasterPublicKey 与 *sm9.EncryptMasterPublicKey
//
// 返回值：
//   - string: 主公钥 ASN.1 编码的 SM3 摘要，HEX 小写；编码失败时返回空串
func masterPublicKeyFingerprint(pub interface {
	MarshalASN1() ([]byte, error)
}) string {
	derData, err := pub.MarshalASN1()
	if err != nil {
		return ""
	}
	sum := sm3.Sum(derData)
	return hex.EncodeToString(sum[:])
}

// kgcKey 返回 KGC 的唯一标识键，与 gotlcp 信任池的判定规则保持一致。
//
// 参数：
//   - districtName: KGC 属地区域名
//   - districtSerial: KGC 序号
//
// 返回值：
//   - string: 唯一标识键
func kgcKey(districtName string, districtSerial int) string {
	return districtName + "\x00" + strconv.Itoa(districtSerial)
}

// validateFilename 校验文件名是否为不含路径分隔符的纯文件名。
//
// 参数：
//   - filename: 待校验的文件名
//
// 返回值：
//   - error: 文件名为空或包含路径分隔符时返回错误
func validateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("文件名不能为空")
	}
	if filepath.Base(filename) != filename || strings.ContainsAny(filename, `/\`) {
		return fmt.Errorf("文件名不能包含路径分隔符: %s", filename)
	}
	return nil
}

// isParamExtension 判断扩展名是否在信任池扫描白名单中。
//
// 参数：
//   - ext: 小写扩展名（含点）
//
// 返回值：
//   - bool: 在白名单中返回 true
func isParamExtension(ext string) bool {
	for _, e := range ParamExtensions {
		if e == ext {
			return true
		}
	}
	return false
}
