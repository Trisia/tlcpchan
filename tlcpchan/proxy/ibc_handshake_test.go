package proxy

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/security"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/keystore"
)

// ibcTestEnv IBC 集成测试环境：一套测试 KGC、两份用户身份及对应的管理器。
type ibcTestEnv struct {
	keyStoreMgr *security.KeyStoreManager // 注册了服务端/客户端 IBC 身份的管理器
	rootCertMgr *security.RootCertManager // 根证书管理器（IBC 场景为空）
	ibcParamMgr *security.IBCParamManager // 含内置测试 KGC 的信任池
	emptyIBCMgr *security.IBCParamManager // 空信任池，用于验证默认拒绝
}

// newIBCTestEnv 在临时目录中构造 IBC 测试环境。
//
// 参数：
//   - t: 测试上下文
//
// 返回值：
//   - *ibcTestEnv: 生成的 KGC、身份文件与管理器
func newIBCTestEnv(t *testing.T) *ibcTestEnv {
	t.Helper()
	dir := t.TempDir()

	kgc, err := certgen.GenerateIBCParams("tlcpchan.local", 1, tlcp.ValidityPeriod{
		NotBefore: time.Now().Add(-time.Hour).Truncate(time.Second),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	})
	if err != nil {
		t.Fatalf("GenerateIBCParams() error = %v", err)
	}

	// 信任池：写入内置测试 KGC 的公共参数
	ibcParamDir := filepath.Join(dir, "ibcparams")
	ibcParamMgr := security.NewIBCParamManager(ibcParamDir)
	if err := ibcParamMgr.Initialize(); err != nil {
		t.Fatalf("IBCParamManager.Initialize() error = %v", err)
	}
	if _, err := ibcParamMgr.Add("tlcpchan-ibc-kgc.pem", kgc.ParamsPEM); err != nil {
		t.Fatalf("IBCParamManager.Add() error = %v", err)
	}

	emptyIBCMgr := security.NewIBCParamManager(filepath.Join(dir, "empty-ibcparams"))
	if err := emptyIBCMgr.Initialize(); err != nil {
		t.Fatalf("空 IBCParamManager.Initialize() error = %v", err)
	}

	// 两份身份：分别作为服务端与客户端
	keyStoreMgr := security.NewKeyStoreManager()
	for _, item := range []struct {
		name     string
		identity string
	}{
		{"ibc-server", "server@tlcpchan.local"},
		{"ibc-client", "client@tlcpchan.local"},
	} {
		generated, err := certgen.GenerateIBCIdentity(kgc.Master, []byte(item.identity))
		if err != nil {
			t.Fatalf("GenerateIBCIdentity(%s) error = %v", item.identity, err)
		}
		params, err := writeIBCIdentityFiles(dir, item.name, generated, kgc.ParamsPEM)
		if err != nil {
			t.Fatalf("写入 IBC 身份文件失败: %v", err)
		}
		if _, err := keyStoreMgr.Create(item.name, keystore.LoaderTypeIBCFile, params, false); err != nil {
			t.Fatalf("创建 IBC keystore %s 失败: %v", item.name, err)
		}
	}

	return &ibcTestEnv{
		keyStoreMgr: keyStoreMgr,
		rootCertMgr: security.NewRootCertManager(dir),
		ibcParamMgr: ibcParamMgr,
		emptyIBCMgr: emptyIBCMgr,
	}
}

// writeIBCIdentityFiles 将一份 IBC 身份写成五个材料文件并返回 ibc-file 参数。
//
// 参数：
//   - dir: 材料文件所在目录
//   - name: keystore 名称，同时作为文件名前缀
//   - generated: 生成好的标识与三把用户私钥（PEM 编码）
//   - paramsPEM: KGC 公共参数 PEM
//
// 返回值：
//   - map[string]string: ibc-file 加载器参数，值为文件绝对路径
//   - error: 写文件失败时返回错误
func writeIBCIdentityFiles(dir, name string, generated *certgen.GeneratedIBCIdentity, paramsPEM []byte) (map[string]string, error) {
	files := []struct {
		key        string
		suffix     string
		data       []byte
		permission os.FileMode
	}{
		{keystore.IBCParamIdentity, "identity.txt", generated.Identity, 0644},
		{keystore.IBCParamParams, "params.pem", paramsPEM, 0644},
		{keystore.IBCParamSignKey, "sign.key", generated.SignKeyPEM, 0600},
		{keystore.IBCParamEncKey, "enc.key", generated.EncKeyPEM, 0600},
		{keystore.IBCParamKexKey, "kex.key", generated.KexKeyPEM, 0600},
	}

	params := make(map[string]string, len(files))
	for _, file := range files {
		path := filepath.Join(dir, name+"-"+file.suffix)
		if err := os.WriteFile(path, file.data, file.permission); err != nil {
			return nil, err
		}
		params[file.key] = path
	}
	return params, nil
}

// ibcInstanceConfig 构造使用指定 IBC 身份与套件的实例配置。
//
// 参数：
//   - instanceType: 实例类型，TypeServer 或 TypeClient
//   - name: 实例名称
//   - ibcKeyStoreName: 按名称引用的 IBC keystore 名称
//   - suite: 单个 TLCP 密码套件名称，如 "IBC_SM4_GCM_SM3"
//
// 返回值：
//   - *config.InstanceConfig: 实例配置
func ibcInstanceConfig(instanceType, name, ibcKeyStoreName, suite string) *config.InstanceConfig {
	return &config.InstanceConfig{
		Name:     name,
		Type:     instanceType,
		Protocol: "tlcp",
		Enabled:  true,
		TLCP: config.TLCPConfig{
			ClientAuthType: "no-client-cert",
			CipherSuites:   []string{suite},
			IBCKeystore: &config.KeyStoreConfig{
				Type:   keystore.LoaderTypeNamed,
				Params: map[string]string{"name": ibcKeyStoreName},
			},
		},
	}
}

// TestIBCAdapterHandshakeAllSuites 验证 4 个 IBC/IBSDH 套件经适配器完成端到端握手与数据往返。
func TestIBCAdapterHandshakeAllSuites(t *testing.T) {
	env := newIBCTestEnv(t)

	suites := []string{"IBC_SM4_GCM_SM3", "IBC_SM4_CBC_SM3", "IBSDH_SM4_GCM_SM3", "IBSDH_SM4_CBC_SM3"}
	for _, suite := range suites {
		t.Run(suite, func(t *testing.T) {
			serverAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
			if err != nil {
				t.Fatalf("创建服务端适配器失败: %v", err)
			}
			clientAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
			if err != nil {
				t.Fatalf("创建客户端适配器失败: %v", err)
			}

			serverCfg := ibcInstanceConfig(TypeServer, "ibc-server-instance", "ibc-server", suite)
			if err := serverAdapter.ReloadConfig(serverCfg); err != nil {
				t.Fatalf("服务端 ReloadConfig() error = %v", err)
			}

			clientCfg := ibcInstanceConfig(TypeClient, "ibc-client-instance", "ibc-client", suite)
			if err := clientAdapter.ReloadConfig(clientCfg); err != nil {
				t.Fatalf("客户端 ReloadConfig() error = %v", err)
			}

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("监听失败: %v", err)
			}
			defer ln.Close()

			tlsLn := serverAdapter.TLCPListener(ln)
			errCh := make(chan error, 1)
			go func() {
				conn, err := tlsLn.Accept()
				if err != nil {
					errCh <- fmt.Errorf("服务端接受连接失败: %w", err)
					return
				}
				defer conn.Close()

				buf := make([]byte, 16)
				n, err := conn.Read(buf)
				if err != nil {
					errCh <- fmt.Errorf("服务端读取失败: %w", err)
					return
				}
				if _, err := conn.Write(append([]byte("echo:"), buf[:n]...)); err != nil {
					errCh <- fmt.Errorf("服务端写入失败: %w", err)
					return
				}
				errCh <- nil
			}()

			conn, err := clientAdapter.DialTLCP("tcp", tlsLn.Addr().String(), clientCfg)
			if err != nil {
				t.Fatalf("客户端握手失败: %v", err)
			}
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

			if _, err := conn.Write([]byte("ping")); err != nil {
				t.Fatalf("客户端写入失败: %v", err)
			}

			buf := make([]byte, 9)
			if _, err := io.ReadFull(conn, buf); err != nil {
				t.Fatalf("客户端读取回显失败: %v", err)
			}
			if string(buf) != "echo:ping" {
				t.Fatalf("回显内容 = %q, want %q", buf, "echo:ping")
			}

			if err := <-errCh; err != nil {
				t.Fatalf("服务端错误: %v", err)
			}
		})
	}
}

// TestIBCAdapterRejectsUntrustedKGC 验证信任池为空时拒绝握手（默认拒绝语义）。
func TestIBCAdapterRejectsUntrustedKGC(t *testing.T) {
	env := newIBCTestEnv(t)

	serverAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
	if err != nil {
		t.Fatalf("创建服务端适配器失败: %v", err)
	}
	clientAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.emptyIBCMgr)
	if err != nil {
		t.Fatalf("创建客户端适配器失败: %v", err)
	}

	serverCfg := ibcInstanceConfig(TypeServer, "ibc-server-instance", "ibc-server", "IBC_SM4_GCM_SM3")
	if err := serverAdapter.ReloadConfig(serverCfg); err != nil {
		t.Fatalf("服务端 ReloadConfig() error = %v", err)
	}

	clientCfg := ibcInstanceConfig(TypeClient, "ibc-client-instance", "ibc-client", "IBC_SM4_GCM_SM3")
	if err := clientAdapter.ReloadConfig(clientCfg); err != nil {
		t.Fatalf("客户端 ReloadConfig() error = %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}
	defer ln.Close()

	tlsLn := serverAdapter.TLCPListener(ln)
	go func() {
		conn, err := tlsLn.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	conn, err := clientAdapter.DialTLCP("tcp", tlsLn.Addr().String(), clientCfg)
	if err == nil {
		conn.Close()
		t.Fatal("信任池为空时握手应当失败，但成功了")
	}
}

// writeTLCPCertificateMaterials 生成一套 TLCP 证书对并写盘，返回 file 加载器参数。
//
// 参数：
//   - t: 测试上下文
//   - dir: 证书文件所在目录
//
// 返回值：
//   - map[string]string: file 加载器参数（绝对路径）
func writeTLCPCertificateMaterials(t *testing.T, dir string) map[string]string {
	t.Helper()
	rootCA, err := certgen.GenerateTLCPRootCA(certgen.CertGenConfig{
		Type:       certgen.CertTypeRootCA,
		CommonName: "tlcpchan-ibc-test-ca",
		Org:        "tlcpchan",
		Years:      10,
	})
	if err != nil {
		t.Fatalf("GenerateTLCPRootCA() error = %v", err)
	}

	caCertPath := filepath.Join(dir, "ca.crt")
	caKeyPath := filepath.Join(dir, "ca.key")
	if err := certgen.SaveCertToFile(rootCA.CertPEM, rootCA.KeyPEM, caCertPath, caKeyPath); err != nil {
		t.Fatalf("保存根 CA 失败: %v", err)
	}

	signerCert, signerKey, err := certgen.LoadTLCPCertFromFile(caCertPath, caKeyPath)
	if err != nil {
		t.Fatalf("LoadTLCPCertFromFile() error = %v", err)
	}
	signCert, encCert, err := certgen.GenerateTLCPPair(signerCert, signerKey,
		certgen.CertGenConfig{
			Type:       certgen.CertTypeTLCPSign,
			CommonName: "tlcpchan-ibc-test-sign",
			Org:        "tlcpchan",
			Years:      5,
		},
		certgen.CertGenConfig{
			Type:       certgen.CertTypeTLCPEnc,
			CommonName: "tlcpchan-ibc-test-enc",
			Org:        "tlcpchan",
			Years:      5,
		})
	if err != nil {
		t.Fatalf("GenerateTLCPPair() error = %v", err)
	}

	signCertPath := filepath.Join(dir, "sign.crt")
	signKeyPath := filepath.Join(dir, "sign.key")
	encCertPath := filepath.Join(dir, "enc.crt")
	encKeyPath := filepath.Join(dir, "enc.key")
	if err := certgen.SaveCertToFile(signCert.CertPEM, signCert.KeyPEM, signCertPath, signKeyPath); err != nil {
		t.Fatalf("保存签名证书失败: %v", err)
	}
	if err := certgen.SaveCertToFile(encCert.CertPEM, encCert.KeyPEM, encCertPath, encKeyPath); err != nil {
		t.Fatalf("保存加密证书失败: %v", err)
	}

	return map[string]string{
		"sign-cert": signCertPath,
		"sign-key":  signKeyPath,
		"enc-cert":  encCertPath,
		"enc-key":   encKeyPath,
	}
}

// TestIBCMixedSuiteNegotiation 验证同一实例并存证书与 IBC 身份时能按对端能力协商到预期套件。
func TestIBCMixedSuiteNegotiation(t *testing.T) {
	env := newIBCTestEnv(t)

	certParams := writeTLCPCertificateMaterials(t, t.TempDir())
	if _, err := env.keyStoreMgr.Create("tlcp-cert", keystore.LoaderTypeFile, certParams, false); err != nil {
		t.Fatalf("创建证书 keystore 失败: %v", err)
	}

	serverAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
	if err != nil {
		t.Fatalf("创建服务端适配器失败: %v", err)
	}

	serverCfg := &config.InstanceConfig{
		Name:     "mixed-server",
		Type:     TypeServer,
		Protocol: "tlcp",
		Enabled:  true,
		TLCP: config.TLCPConfig{
			ClientAuthType: "no-client-cert",
			CipherSuites:   []string{"ECC_SM4_GCM_SM3", "IBC_SM4_GCM_SM3"},
			Keystore:       &config.KeyStoreConfig{Type: keystore.LoaderTypeNamed, Params: map[string]string{"name": "tlcp-cert"}},
			IBCKeystore:    &config.KeyStoreConfig{Type: keystore.LoaderTypeNamed, Params: map[string]string{"name": "ibc-server"}},
		},
	}
	if err := serverAdapter.ReloadConfig(serverCfg); err != nil {
		t.Fatalf("混合实例 ReloadConfig() error = %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听失败: %v", err)
	}
	defer ln.Close()

	tlsLn := serverAdapter.TLCPListener(ln)
	go func() {
		for {
			conn, err := tlsLn.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 16)
				n, err := c.Read(buf)
				if err != nil {
					return
				}
				_, _ = c.Write(buf[:n])
			}(conn)
		}
	}()

	tests := []struct {
		name         string
		ibc          bool
		clientSuites []string
		wantSuite    uint16
	}{
		{"证书客户端协商到 ECC 套件", false, []string{"ECC_SM4_GCM_SM3", "IBC_SM4_GCM_SM3"}, tlcp.ECC_SM4_GCM_SM3},
		{"IBC 客户端协商到 IBC 套件", true, []string{"IBC_SM4_GCM_SM3"}, tlcp.IBC_SM4_GCM_SM3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientCfg := &config.InstanceConfig{
				Name:     "mixed-client",
				Type:     TypeClient,
				Protocol: "tlcp",
				Enabled:  true,
				TLCP: config.TLCPConfig{
					ClientAuthType: "no-client-cert",
					// IBC 客户端只列 IBC 套件，以验证服务端在同一实例上仍能协商 IBC
					CipherSuites:       tt.clientSuites,
					InsecureSkipVerify: !tt.ibc, // 证书客户端跳过证书校验，仅验证套件协商
				},
			}
			if tt.ibc {
				clientCfg.TLCP.IBCKeystore = &config.KeyStoreConfig{
					Type:   keystore.LoaderTypeNamed,
					Params: map[string]string{"name": "ibc-client"},
				}
			} else {
				clientCfg.TLCP.Keystore = &config.KeyStoreConfig{
					Type:   keystore.LoaderTypeNamed,
					Params: map[string]string{"name": "tlcp-cert"},
				}
			}

			clientAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
			if err != nil {
				t.Fatalf("创建客户端适配器失败: %v", err)
			}
			if err := clientAdapter.ReloadConfig(clientCfg); err != nil {
				t.Fatalf("客户端 ReloadConfig() error = %v", err)
			}

			conn, err := clientAdapter.DialTLCP("tcp", tlsLn.Addr().String(), clientCfg)
			if err != nil {
				t.Fatalf("握手失败: %v", err)
			}
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

			if _, err := conn.Write([]byte("suite")); err != nil {
				t.Fatalf("写入失败: %v", err)
			}
			buf := make([]byte, 5)
			if _, err := io.ReadFull(conn, buf); err != nil {
				t.Fatalf("读取回显失败: %v", err)
			}

			state := conn.(*tlcp.Conn).ConnectionState()
			if state.CipherSuite != tt.wantSuite {
				t.Fatalf("协商套件 = 0x%04X (%s), want 0x%04X (%s)",
					state.CipherSuite, tlcp.CipherSuiteName(state.CipherSuite),
					tt.wantSuite, tlcp.CipherSuiteName(tt.wantSuite))
			}
		})
	}
}

// TestIBCMutualAuthentication 验证纯 IBC 服务端（无证书）按 client-auth-type 要求客户端身份。
func TestIBCMutualAuthentication(t *testing.T) {
	suites := []string{"IBSDH_SM4_GCM_SM3", "IBC_SM4_GCM_SM3"}
	for _, suite := range suites {
		t.Run(suite, func(t *testing.T) {
			env := newIBCTestEnv(t)

			serverAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
			if err != nil {
				t.Fatalf("创建服务端适配器失败: %v", err)
			}
			serverCfg := ibcInstanceConfig(TypeServer, "mutual-server", "ibc-server", suite)
			serverCfg.TLCP.ClientAuthType = "require-and-verify-client-cert"
			if err := serverAdapter.ReloadConfig(serverCfg); err != nil {
				t.Fatalf("服务端 ReloadConfig() error = %v", err)
			}

			clientAdapter, err := NewTLCPAdapter(env.keyStoreMgr, env.rootCertMgr, env.ibcParamMgr)
			if err != nil {
				t.Fatalf("创建客户端适配器失败: %v", err)
			}
			clientCfg := ibcInstanceConfig(TypeClient, "mutual-client", "ibc-client", suite)
			if err := clientAdapter.ReloadConfig(clientCfg); err != nil {
				t.Fatalf("客户端 ReloadConfig() error = %v", err)
			}

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("监听失败: %v", err)
			}
			defer ln.Close()

			tlsLn := serverAdapter.TLCPListener(ln)
			errCh := make(chan error, 1)
			go func() {
				conn, err := tlsLn.Accept()
				if err != nil {
					errCh <- err
					return
				}
				defer conn.Close()

				// Accept 返回的连接尚未握手，先显式握手再检查客户端身份
				if err := conn.(*tlcp.Conn).Handshake(); err != nil {
					errCh <- fmt.Errorf("服务端握手失败: %w", err)
					return
				}

				// 服务端应能看到客户端标识与已校验的客户端公共参数
				state := conn.(*tlcp.Conn).ConnectionState()
				if string(state.PeerIBCIdentity) != "client@tlcpchan.local" {
					errCh <- fmt.Errorf("客户端标识 = %q, want client@tlcpchan.local", state.PeerIBCIdentity)
					return
				}
				if state.PeerIBCSysParams == nil {
					errCh <- fmt.Errorf("服务端未获得客户端 KGC 公共参数")
					return
				}

				buf := make([]byte, 16)
				n, err := conn.Read(buf)
				if err != nil {
					errCh <- err
					return
				}
				_, err = conn.Write(buf[:n])
				errCh <- err
			}()

			conn, err := clientAdapter.DialTLCP("tcp", tlsLn.Addr().String(), clientCfg)
			if err != nil {
				select {
				case serverErr := <-errCh:
					t.Fatalf("双向认证握手失败: %v（服务端: %v）", err, serverErr)
				default:
					t.Fatalf("双向认证握手失败: %v", err)
				}
			}
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

			if _, err := conn.Write([]byte("mutual")); err != nil {
				t.Fatalf("写入失败: %v", err)
			}
			buf := make([]byte, 6)
			if _, err := io.ReadFull(conn, buf); err != nil {
				t.Fatalf("读取回显失败: %v", err)
			}
			if string(buf) != "mutual" {
				t.Fatalf("回显 = %q", buf)
			}

			state := conn.(*tlcp.Conn).ConnectionState()
			if state.CipherSuite != tlcp.IBSDH_SM4_GCM_SM3 && state.CipherSuite != tlcp.IBC_SM4_GCM_SM3 {
				t.Fatalf("协商套件 = %s", tlcp.CipherSuiteName(state.CipherSuite))
			}

			if err := <-errCh; err != nil {
				t.Fatalf("服务端错误: %v", err)
			}
		})
	}
}
