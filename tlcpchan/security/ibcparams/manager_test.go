package ibcparams

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/security/certgen"
)

// newTestKGC 生成一套测试 KGC 公共参数。
//
// 参数：
//   - t: 测试上下文
//   - districtName: KGC 属地区域名
//   - districtSerial: KGC 序号
//
// 返回值：
//   - *certgen.GeneratedIBCParams: 生成的公共参数与主密钥
func newTestKGC(t *testing.T, districtName string, districtSerial int) *certgen.GeneratedIBCParams {
	t.Helper()
	kgc, err := certgen.GenerateIBCParams(districtName, districtSerial, tlcp.ValidityPeriod{
		NotBefore: time.Now().Add(-time.Hour).Truncate(time.Second),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}
	return kgc
}

// TestManagerAddAndList 验证添加公共参数后元信息与信任池都正确更新。
func TestManagerAddAndList(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if len(mgr.List()) != 0 {
		t.Fatalf("初始信任池应为空，实际 %d 条", len(mgr.List()))
	}

	kgc := newTestKGC(t, "tlcpchan.local", 1)
	param, err := mgr.Add("tlcpchan-ibc-kgc.pem", kgc.ParamsPEM)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if param.Filename != "tlcpchan-ibc-kgc.pem" {
		t.Fatalf("Filename = %q", param.Filename)
	}
	if param.DistrictName != "tlcpchan.local" || param.DistrictSerial != 1 {
		t.Fatalf("KGC 信息 = %s#%d, want tlcpchan.local#1", param.DistrictName, param.DistrictSerial)
	}
	if param.SignKeyFingerprint == "" || param.EncKeyFingerprint == "" {
		t.Fatal("主公钥指纹不应为空")
	}
	if param.NotAfter.IsZero() {
		t.Fatal("NotAfter 不应为零值")
	}

	// 信任池应命中该 KGC
	if !mgr.GetPool().Contains(kgc.Params) {
		t.Fatal("信任池未命中已添加的 KGC 公共参数")
	}

	list := mgr.List()
	if len(list) != 1 {
		t.Fatalf("List() 长度 = %d, want 1", len(list))
	}

	// 文件权限应为 0600
	info, err := os.Stat(filepath.Join(dir, "tlcpchan-ibc-kgc.pem"))
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("文件权限 = %v, want 0600", info.Mode().Perm())
	}

	// Get 与 ReadFile 应可读取
	if _, err := mgr.Get("tlcpchan-ibc-kgc.pem"); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if data, err := mgr.ReadFile("tlcpchan-ibc-kgc.pem"); err != nil || len(data) == 0 {
		t.Fatalf("ReadFile() = %d 字节, error = %v", len(data), err)
	}
}

// TestManagerAddErrors 验证文件名非法、数据非法与同一 KGC 重复添加都会被拒绝。
func TestManagerAddErrors(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	kgc := newTestKGC(t, "tlcpchan.local", 1)
	if _, err := mgr.Add("tlcpchan-ibc-kgc.pem", kgc.ParamsPEM); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 同一 KGC 换个文件名再次添加应被拒绝
	another := newTestKGC(t, "other.local", 2)
	tests := []struct {
		name     string
		filename string
		data     []byte
	}{
		{"文件名为空", "", kgc.ParamsPEM},
		{"文件名含路径分隔符", "sub/kgc.pem", kgc.ParamsPEM},
		{"扩展名不在白名单", "kgc.txt", kgc.ParamsPEM},
		{"数据为空", "empty.pem", nil},
		{"数据格式非法", "invalid.pem", []byte("not-ibc-params")},
		{"同一 KGC 重复添加", "duplicate.pem", kgc.ParamsPEM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := mgr.Add(tt.filename, tt.data); err == nil {
				t.Fatalf("Add(%q) 应当返回错误", tt.filename)
			}
		})
	}

	// 不同 KGC 应可正常添加
	if _, err := mgr.Add("other.pem", another.ParamsPEM); err != nil {
		t.Fatalf("添加不同 KGC 失败: %v", err)
	}
	if len(mgr.List()) != 2 {
		t.Fatalf("List() 长度 = %d, want 2", len(mgr.List()))
	}
}

// TestManagerDeleteAndReload 验证删除与重载后的状态一致性。
func TestManagerDeleteAndReload(t *testing.T) {
	dir := t.TempDir()
	mgr := NewManager(dir)
	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	kgc := newTestKGC(t, "tlcpchan.local", 1)
	if _, err := mgr.Add("kgc.pem", kgc.ParamsPEM); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// 从磁盘手工放入一个参数文件后重载，应被扫描并加入信任池
	second := newTestKGC(t, "example.local", 5)
	if err := os.WriteFile(filepath.Join(dir, "second.pem"), second.ParamsPEM, 0600); err != nil {
		t.Fatalf("写入第二个参数文件失败: %v", err)
	}
	if err := mgr.Reload(); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
	if len(mgr.List()) != 2 {
		t.Fatalf("Reload() 后 List() 长度 = %d, want 2", len(mgr.List()))
	}

	// 删除后信任池不应再命中
	if err := mgr.Delete("kgc.pem"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(mgr.List()) != 1 {
		t.Fatalf("Delete() 后 List() 长度 = %d, want 1", len(mgr.List()))
	}
	if mgr.GetPool().Contains(kgc.Params) {
		t.Fatal("删除后信任池仍命中被删除的 KGC")
	}

	// 未提供扩展名的文件不参与扫描
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), kgc.ParamsPEM, 0600); err != nil {
		t.Fatalf("写入非法扩展名文件失败: %v", err)
	}
	if err := mgr.Reload(); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
	if len(mgr.List()) != 1 {
		t.Fatalf("非法扩展名文件被扫描，List() 长度 = %d, want 1", len(mgr.List()))
	}

	// 子目录不参与扫描（信任池只接受顶层文件）
	subDir := filepath.Join(dir, "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested.pem"), second.ParamsPEM, 0600); err != nil {
		t.Fatalf("写入子目录文件失败: %v", err)
	}
	if err := mgr.Reload(); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
	if len(mgr.List()) != 1 {
		t.Fatalf("子目录被扫描，List() 长度 = %d, want 1", len(mgr.List()))
	}
}

// TestManagerGetNotFound 验证查询不存在的条目返回错误。
func TestManagerGetNotFound(t *testing.T) {
	mgr := NewManager(t.TempDir())
	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if _, err := mgr.Get("missing.pem"); err == nil {
		t.Fatal("Get(不存在的条目) 应当返回错误")
	}
	if _, err := mgr.ReadFile("sub/missing.pem"); err == nil {
		t.Fatal("ReadFile(含路径分隔符) 应当返回错误")
	}
}

// TestManagerEmptyBaseDir 验证未配置目录时信任池保持为空且不报错。
func TestManagerEmptyBaseDir(t *testing.T) {
	mgr := NewManager("")
	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if len(mgr.List()) != 0 {
		t.Fatalf("List() 长度 = %d, want 0", len(mgr.List()))
	}
	if mgr.GetPool() == nil {
		t.Fatal("GetPool() 不应返回 nil")
	}
}
