package initialization

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// TestInitializeCreatesIBCMaterials 验证首次初始化会生成内置测试 KGC、预置 IBC 身份并登记配置。
func TestInitializeCreatesIBCMaterials(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.WorkDir = dir
	configPath := filepath.Join(dir, "config.yaml")
	config.Init(cfg, configPath)

	mgr := NewManager(cfg, configPath, dir)
	if mgr.CheckInitialized() {
		t.Fatal("空目录不应被认为已完成初始化")
	}

	if err := mgr.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// 1. KGC 公共参数与主密钥文件
	kgcFiles := []struct {
		path string
		perm os.FileMode
	}{
		{filepath.Join(dir, "ibcparams", "tlcpchan-ibc-kgc.pem"), 0644},
		{filepath.Join(dir, "keystores", "tlcpchan-ibc-kgc-master.key"), 0600},
	}
	for _, file := range kgcFiles {
		info, err := os.Stat(file.path)
		if err != nil {
			t.Fatalf("初始化未生成 %s: %v", file.path, err)
		}
		if info.Mode().Perm() != file.perm {
			t.Fatalf("%s 权限 = %v, want %v", file.path, info.Mode().Perm(), file.perm)
		}
	}

	// 2. 两份预置身份的五项材料
	identityFiles := []struct {
		suffix string
		perm   os.FileMode
	}{
		{"identity.txt", 0644},
		{"params.pem", 0644},
		{"sign.key", 0600},
		{"enc.key", 0600},
		{"kex.key", 0600},
	}
	for _, name := range []string{"default-ibc-server", "default-ibc-client"} {
		for _, file := range identityFiles {
			path := filepath.Join(dir, "keystores", name+"-"+file.suffix)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("初始化未生成 %s: %v", path, err)
			}
			if info.Mode().Perm() != file.perm {
				t.Fatalf("%s 权限 = %v, want %v", path, info.Mode().Perm(), file.perm)
			}
		}
	}

	// 3. 配置中登记了 ibc-file 类型的两个 keystore
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	expected := map[string]string{
		"default-ibc-server": "server@tlcpchan.local",
		"default-ibc-client": "client@tlcpchan.local",
	}
	found := make(map[string]bool)
	for _, ks := range loaded.KeyStores {
		identity, ok := expected[ks.Name]
		if !ok {
			continue
		}
		if ks.Type != keystore.LoaderTypeIBCFile {
			t.Fatalf("keystore %s 类型 = %q, want %q", ks.Name, ks.Type, keystore.LoaderTypeIBCFile)
		}
		for _, key := range []string{
			keystore.IBCParamIdentity,
			keystore.IBCParamParams,
			keystore.IBCParamSignKey,
			keystore.IBCParamEncKey,
			keystore.IBCParamKexKey,
		} {
			if ks.Params[key] == "" {
				t.Fatalf("keystore %s 缺少参数 %s", ks.Name, key)
			}
		}

		// 标识文件内容应与配置的名称对应
		identityPath := filepath.Join(dir, ks.Params[keystore.IBCParamIdentity])
		data, err := os.ReadFile(identityPath)
		if err != nil {
			t.Fatalf("读取标识文件 %s 失败: %v", identityPath, err)
		}
		if string(data) != identity {
			t.Fatalf("keystore %s 标识 = %q, want %q", ks.Name, data, identity)
		}
		found[ks.Name] = true
	}
	if len(found) != len(expected) {
		t.Fatalf("配置中预置 IBC keystore = %v, want %v", found, expected)
	}

	// 4. 信任池可直接装载 KGC 公共参数
	ibcParamMgr := security.NewIBCParamManager(filepath.Join(dir, "ibcparams"))
	if err := ibcParamMgr.Initialize(); err != nil {
		t.Fatalf("IBCParamManager.Initialize() error = %v", err)
	}
	params := ibcParamMgr.List()
	if len(params) != 1 {
		t.Fatalf("信任池条目数 = %d, want 1", len(params))
	}
	if params[0].DistrictName != "tlcpchan.local" || params[0].DistrictSerial != 1 {
		t.Fatalf("信任池 KGC = %s#%d, want tlcpchan.local#1", params[0].DistrictName, params[0].DistrictSerial)
	}
	if params[0].SignKeyFingerprint == "" || params[0].EncKeyFingerprint == "" {
		t.Fatal("信任池主公钥指纹不应为空")
	}

	// 5. 初始化完成后检查通过
	if !mgr.CheckInitialized() {
		t.Fatal("初始化完成后 CheckInitialized() 应为 true")
	}

	// 6. IBC 相关文件缺失不影响已初始化判定（旧配置无需迁移）
	if err := os.RemoveAll(filepath.Join(dir, "ibcparams")); err != nil {
		t.Fatalf("删除 ibcparams 目录失败: %v", err)
	}
	if !mgr.CheckInitialized() {
		t.Fatal("删除 IBC 信任池后 CheckInitialized() 仍应为 true")
	}
}
