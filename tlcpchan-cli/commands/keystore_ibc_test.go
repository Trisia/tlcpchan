package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Trisia/tlcpchan-cli/client"
)

// writeTempMaterial 在临时目录写入材料文件并返回其路径
// 参数：
//   - t: 测试上下文
//   - name: 文件名
//   - content: 文件内容
//
// 返回：
//   - string: 文件绝对路径
func writeTempMaterial(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("写入临时材料 %s 失败: %v", name, err)
	}
	return path
}

// withTestClient 将全局 CLI 客户端临时指向测试服务端
// 参数：
//   - t: 测试上下文
//   - url: 测试服务端地址
//
// 返回：
//   - func(): 恢复原客户端的清理函数
func withTestClient(t *testing.T, url string) func() {
	t.Helper()
	original := cli
	cli = client.NewClient(url)
	return func() { cli = original }
}

// TestKeyStoreCreateIBCFileFields 验证 ibc-file 创建请求的 multipart 字段名与文件内容
func TestKeyStoreCreateIBCFileFields(t *testing.T) {
	identityPath := writeTempMaterial(t, "identity.txt", "server@tlcpchan.local")
	paramsPath := writeTempMaterial(t, "params.pem", "PARAMS")
	signPath := writeTempMaterial(t, "sign.key", "SIGN")
	encPath := writeTempMaterial(t, "enc.key", "ENC")
	kexPath := writeTempMaterial(t, "kex.key", "KEX")

	received := map[string]string{}
	var loaderType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security/keystores" {
			t.Errorf("请求路径错误: %s", r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("解析 multipart 失败: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		loaderType = r.FormValue("loaderType")

		for _, field := range []string{"identity", "params", "signKey", "encKey", "kexKey"} {
			file, _, err := r.FormFile(field)
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(file)
			file.Close()
			received[field] = string(data)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"default-ibc-server","type":"ibc","loaderType":"ibc-file"}`))
	}))
	defer server.Close()
	defer withTestClient(t, server.URL)()

	err := keyStoreCreate([]string{
		"--name", "default-ibc-server",
		"--loader-type", "ibc-file",
		"--identity", identityPath,
		"--params", paramsPath,
		"--sign-key", signPath,
		"--enc-key", encPath,
		"--kex-key", kexPath,
	})
	if err != nil {
		t.Fatalf("keyStoreCreate() 失败: %v", err)
	}

	if loaderType != "ibc-file" {
		t.Errorf("loaderType 字段应为 ibc-file, 实际为 %s", loaderType)
	}

	expected := map[string]string{
		"identity": "server@tlcpchan.local",
		"params":   "PARAMS",
		"signKey":  "SIGN",
		"encKey":   "ENC",
		"kexKey":   "KEX",
	}
	if len(received) != len(expected) {
		t.Errorf("multipart 文件字段数量应为 %d, 实际为 %d (%v)", len(expected), len(received), received)
	}
	for field, content := range expected {
		if received[field] != content {
			t.Errorf("字段 %s 内容应为 %q, 实际为 %q", field, content, received[field])
		}
	}
}

// TestKeyStoreCreateIBCFileRequiresIdentity 验证 ibc-file 缺少标识时直接报错
func TestKeyStoreCreateIBCFileRequiresIdentity(t *testing.T) {
	err := keyStoreCreate([]string{"--name", "ibc-1", "--loader-type", "ibc-file"})
	if err == nil {
		t.Fatal("缺少 --identity 时应返回错误")
	}
}

// TestKeyStoreCreateIBCFileRejectsCertOptions 验证 ibc-file 拒绝证书材料选项
func TestKeyStoreCreateIBCFileRejectsCertOptions(t *testing.T) {
	identityPath := writeTempMaterial(t, "identity.txt", "server@tlcpchan.local")
	certPath := writeTempMaterial(t, "sign.crt", "CERT")

	err := keyStoreCreate([]string{
		"--name", "ibc-1",
		"--loader-type", "ibc-file",
		"--identity", identityPath,
		"--sign-cert", certPath,
	})
	if err == nil {
		t.Fatal("ibc-file 携带 --sign-cert 时应返回错误")
	}
}

// TestKeyStoreUploadIBCFileFields 验证 ibc-file 上传请求的路径与 multipart 字段名
func TestKeyStoreUploadIBCFileFields(t *testing.T) {
	identityPath := writeTempMaterial(t, "identity.txt", "client@tlcpchan.local")
	kexPath := writeTempMaterial(t, "kex.key", "KEX")

	received := map[string]string{}
	var requestPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/security/keystores/default-ibc-client/instances" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("[]"))
			return
		}
		requestPath = r.URL.Path

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("解析 multipart 失败: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, field := range []string{"identity", "params", "signKey", "encKey", "kexKey"} {
			file, _, err := r.FormFile(field)
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(file)
			file.Close()
			received[field] = string(data)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"default-ibc-client","type":"ibc","loaderType":"ibc-file"}`))
	}))
	defer server.Close()
	defer withTestClient(t, server.URL)()

	// 不显式指定 --loader-type，由 IBC 专有参数自动判定
	err := keyStoreUploadCertificates([]string{
		"default-ibc-client",
		"--identity", identityPath,
		"--kex-key", kexPath,
	})
	if err != nil {
		t.Fatalf("keyStoreUploadCertificates() 失败: %v", err)
	}

	if requestPath != "/api/security/keystores/default-ibc-client/upload" {
		t.Errorf("上传路径错误: %s", requestPath)
	}
	if received["identity"] != "client@tlcpchan.local" || received["kexKey"] != "KEX" {
		t.Errorf("上传字段内容错误: %v", received)
	}
	if len(received) != 2 {
		t.Errorf("上传字段数量应为 2, 实际为 %d (%v)", len(received), received)
	}
}

// TestKeyStoreUploadIBCFileRejectsCertOptions 验证 ibc-file 上传拒绝证书材料选项
func TestKeyStoreUploadIBCFileRejectsCertOptions(t *testing.T) {
	certPath := writeTempMaterial(t, "sign.crt", "CERT")

	err := keyStoreUploadCertificates([]string{
		"ibc-1",
		"--loader-type", "ibc-file",
		"--sign-cert", certPath,
	})
	if err == nil {
		t.Fatal("ibc-file 上传携带 --sign-cert 时应返回错误")
	}
}

// TestKeyStoreUpdateIBCParams 验证 ibc-file 更新的参数键名
func TestKeyStoreUpdateIBCParams(t *testing.T) {
	var params map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/security/keystores/default-ibc-server/instances" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("[]"))
			return
		}

		var body struct {
			Params map[string]string `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("解析请求体失败: %v", err)
			return
		}
		params = body.Params

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"default-ibc-server","type":"ibc","loaderType":"ibc-file"}`))
	}))
	defer server.Close()
	defer withTestClient(t, server.URL)()

	err := keyStoreUpdateParams([]string{
		"default-ibc-server",
		"--identity", "./keystores/default-ibc-server-identity.txt",
		"--params", "./keystores/default-ibc-server-params.pem",
		"--sign-key", "./keystores/default-ibc-server-sign.key",
		"--enc-key", "./keystores/default-ibc-server-enc.key",
		"--kex-key", "./keystores/default-ibc-server-kex.key",
	})
	if err != nil {
		t.Fatalf("keyStoreUpdateParams() 失败: %v", err)
	}

	expected := map[string]string{
		"identity": "./keystores/default-ibc-server-identity.txt",
		"params":   "./keystores/default-ibc-server-params.pem",
		"sign-key": "./keystores/default-ibc-server-sign.key",
		"enc-key":  "./keystores/default-ibc-server-enc.key",
		"kex-key":  "./keystores/default-ibc-server-kex.key",
	}
	if len(params) != len(expected) {
		t.Fatalf("参数数量应为 %d, 实际为 %d (%v)", len(expected), len(params), params)
	}
	for key, value := range expected {
		if params[key] != value {
			t.Errorf("参数 %s 应为 %s, 实际为 %s", key, value, params[key])
		}
	}
}

// TestKeyStoreGenerateIBC 验证 ibc 类型生成请求的 JSON 请求体
func TestKeyStoreGenerateIBC(t *testing.T) {
	var body map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security/keystores/generate" {
			t.Errorf("请求路径错误: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("解析请求体失败: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"default-ibc-server","type":"ibc","loaderType":"ibc-file"}`))
	}))
	defer server.Close()
	defer withTestClient(t, server.URL)()

	err := keyStoreGenerate([]string{
		"--name", "default-ibc-server",
		"--type", "ibc",
		"--identity", "server@tlcpchan.local",
		"--district-name", "tlcpchan.local",
		"--district-serial", "1",
	})
	if err != nil {
		t.Fatalf("keyStoreGenerate() 失败: %v", err)
	}

	if body["type"] != "ibc" || body["identity"] != "server@tlcpchan.local" {
		t.Errorf("请求体字段错误: %v", body)
	}
	if body["districtName"] != "tlcpchan.local" || body["districtSerial"] != float64(1) {
		t.Errorf("KGC 定位字段错误: %v", body)
	}
	if _, ok := body["certConfig"]; ok {
		t.Errorf("ibc 类型请求不应携带 certConfig: %v", body)
	}
}

// TestKeyStoreGenerateIBCRequiresIdentity 验证 ibc 类型缺少标识时直接报错
func TestKeyStoreGenerateIBCRequiresIdentity(t *testing.T) {
	err := keyStoreGenerate([]string{"--name", "ibc-1", "--type", "ibc"})
	if err == nil {
		t.Fatal("ibc 类型缺少 --identity 时应返回错误")
	}
}

// TestKeyStoreGenerateCertificateStillWorks 验证证书类型生成请求仍携带 certConfig
func TestKeyStoreGenerateCertificateStillWorks(t *testing.T) {
	var body map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("解析请求体失败: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"ks-1","type":"tlcp","loaderType":"ibc-file"}`))
	}))
	defer server.Close()
	defer withTestClient(t, server.URL)()

	err := keyStoreGenerate([]string{"--name", "ks-1", "--type", "tlcp", "--cn", "tlcpchan.local"})
	if err != nil {
		t.Fatalf("keyStoreGenerate() 失败: %v", err)
	}

	certConfig, ok := body["certConfig"].(map[string]interface{})
	if !ok {
		t.Fatalf("证书类型请求应携带 certConfig: %v", body)
	}
	if certConfig["commonName"] != "tlcpchan.local" {
		t.Errorf("commonName 错误: %v", certConfig["commonName"])
	}
}
