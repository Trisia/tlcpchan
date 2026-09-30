package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestClient_ListIBCParams 验证 IBC 信任池列表接口的路径与解析
func TestClient_ListIBCParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security/ibcparams" {
			t.Errorf("请求路径应为 /api/security/ibcparams, 实际为 %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("请求方法应为 GET, 实际为 %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"filename":"tlcpchan-ibc-kgc.pem","districtName":"tlcpchan.local","districtSerial":1,
			"notBefore":"2025-01-01T00:00:00Z","notAfter":"2035-01-01T00:00:00Z","issuerIdentity":"kgc@tlcpchan.local",
			"signKeyFingerprint":"aabb","encKeyFingerprint":"ccdd"}]`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	params, err := client.ListIBCParams()
	if err != nil {
		t.Fatalf("ListIBCParams() 失败: %v", err)
	}
	if len(params) != 1 {
		t.Fatalf("期望 1 个条目, 实际 %d", len(params))
	}
	if params[0].Filename != "tlcpchan-ibc-kgc.pem" {
		t.Errorf("文件名错误: %s", params[0].Filename)
	}
	if params[0].DistrictName != "tlcpchan.local" || params[0].DistrictSerial != 1 {
		t.Errorf("KGC 标识错误: %s#%d", params[0].DistrictName, params[0].DistrictSerial)
	}
	if params[0].SignKeyFingerprint != "aabb" || params[0].EncKeyFingerprint != "ccdd" {
		t.Errorf("主公钥指纹解析错误: %s / %s", params[0].SignKeyFingerprint, params[0].EncKeyFingerprint)
	}
}

// TestClient_ListIBCParams_Empty 验证信任池为空时返回空列表
func TestClient_ListIBCParams_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	params, err := client.ListIBCParams()
	if err != nil {
		t.Fatalf("ListIBCParams() 失败: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("期望空列表, 实际 %d 个条目", len(params))
	}
}

// TestClient_AddIBCParam 验证添加 KGC 公共参数的 multipart 字段名与内容
func TestClient_AddIBCParam(t *testing.T) {
	const paramsContent = "-----BEGIN IBC PARAMETERS-----\ntest\n-----END IBC PARAMETERS-----\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security/ibcparams" {
			t.Errorf("请求路径应为 /api/security/ibcparams, 实际为 %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("请求方法应为 POST, 实际为 %s", r.Method)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("解析 multipart 失败: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if got := r.FormValue("filename"); got != "my-kgc.pem" {
			t.Errorf("filename 字段应为 my-kgc.pem, 实际为 %s", got)
		}
		file, _, err := r.FormFile("params")
		if err != nil {
			t.Errorf("缺少 params 文件字段: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, _ := io.ReadAll(file)
		if string(data) != paramsContent {
			t.Errorf("params 内容不匹配: %s", string(data))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"filename":"my-kgc.pem","districtName":"tlcpchan.local","districtSerial":1}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	param, err := client.AddIBCParam("my-kgc.pem", []byte(paramsContent))
	if err != nil {
		t.Fatalf("AddIBCParam() 失败: %v", err)
	}
	if param.Filename != "my-kgc.pem" {
		t.Errorf("返回文件名错误: %s", param.Filename)
	}
}

// TestClient_DownloadIBCParam 验证下载路径转义
func TestClient_DownloadIBCParam(t *testing.T) {
	const content = "params-bytes"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security/ibcparams/my-kgc.pem" {
			t.Errorf("请求路径错误: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	data, err := client.DownloadIBCParam("my-kgc.pem")
	if err != nil {
		t.Fatalf("DownloadIBCParam() 失败: %v", err)
	}
	if string(data) != content {
		t.Errorf("下载内容错误: %s", string(data))
	}
}

// TestClient_DeleteIBCParam 验证删除请求的方法与路径
func TestClient_DeleteIBCParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("请求方法应为 DELETE, 实际为 %s", r.Method)
		}
		if r.URL.Path != "/api/security/ibcparams/my-kgc.pem" {
			t.Errorf("请求路径错误: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.DeleteIBCParam("my-kgc.pem"); err != nil {
		t.Fatalf("DeleteIBCParam() 失败: %v", err)
	}
}

// TestClient_GenerateIBCParam 验证生成测试 KGC 的 JSON 请求体
func TestClient_GenerateIBCParam(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security/ibcparams/generate" {
			t.Errorf("请求路径错误: %s", r.URL.Path)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("解析请求体失败: %v", err)
			return
		}
		if body["districtName"] != "tlcpchan.local" {
			t.Errorf("districtName 错误: %v", body["districtName"])
		}
		if body["districtSerial"] != float64(2) {
			t.Errorf("districtSerial 错误: %v", body["districtSerial"])
		}
		if body["years"] != float64(5) {
			t.Errorf("years 错误: %v", body["years"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"filename":"tlcpchan-ibc-kgc-2.pem","districtName":"tlcpchan.local","districtSerial":2}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	param, err := client.GenerateIBCParam(GenerateIBCParamRequest{
		DistrictName:   "tlcpchan.local",
		DistrictSerial: 2,
		Years:          5,
	})
	if err != nil {
		t.Fatalf("GenerateIBCParam() 失败: %v", err)
	}
	if param.Filename != "tlcpchan-ibc-kgc-2.pem" {
		t.Errorf("返回文件名错误: %s", param.Filename)
	}
}

// TestClient_ReloadIBCParams 验证信任池重载请求
func TestClient_ReloadIBCParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("请求方法应为 POST, 实际为 %s", r.Method)
		}
		if r.URL.Path != "/api/security/ibcparams/reload" {
			t.Errorf("请求路径错误: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.ReloadIBCParams(); err != nil {
		t.Fatalf("ReloadIBCParams() 失败: %v", err)
	}
}

// TestGenerateKeyStoreRequest_IBC 验证 ibc 类型生成请求的 JSON 字段符合后端契约
// 要求：只包含 name/type/protected/identity/districtName/districtSerial，不出现 certConfig
func TestGenerateKeyStoreRequest_IBC(t *testing.T) {
	serial := 1
	req := GenerateKeyStoreRequest{
		Name:           "default-ibc-server",
		Type:           "ibc",
		Protected:      false,
		Identity:       "server@tlcpchan.local",
		DistrictName:   "tlcpchan.local",
		DistrictSerial: &serial,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	if body["type"] != "ibc" || body["name"] != "default-ibc-server" {
		t.Errorf("name/type 字段错误: %v", body)
	}
	if body["identity"] != "server@tlcpchan.local" {
		t.Errorf("identity 字段错误: %v", body["identity"])
	}
	if body["districtName"] != "tlcpchan.local" || body["districtSerial"] != float64(1) {
		t.Errorf("KGC 定位字段错误: %v", body)
	}
	if _, ok := body["certConfig"]; ok {
		t.Errorf("ibc 类型请求不应携带 certConfig: %s", string(data))
	}
}

// TestKeyStoreInfo_IBC 验证 keystore 响应中的 ibc 元信息解析
func TestKeyStoreInfo_IBC(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"default-ibc-server","type":"ibc","loaderType":"ibc-file","protected":false,
			"createdAt":"2025-01-01T00:00:00Z","updatedAt":"2025-01-01T00:00:00Z",
			"ibc":{"identity":"server@tlcpchan.local","hasParams":true,"districtName":"tlcpchan.local","districtSerial":1,
			"notBefore":"2025-01-01T00:00:00Z","notAfter":"2035-01-01T00:00:00Z",
			"hasSignKey":true,"hasEncryptKey":true,"hasKeyExchangeKey":false}}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ks, err := client.GetKeyStore("default-ibc-server")
	if err != nil {
		t.Fatalf("GetKeyStore() 失败: %v", err)
	}
	if ks.Type != "ibc" || ks.LoaderType != "ibc-file" {
		t.Errorf("类型字段错误: %s / %s", ks.Type, ks.LoaderType)
	}
	if ks.IBC == nil {
		t.Fatal("ibc 元信息未解析")
	}
	if ks.IBC.Identity != "server@tlcpchan.local" || !ks.IBC.HasParams || !ks.IBC.HasSignKey {
		t.Errorf("ibc 元信息内容错误: %+v", *ks.IBC)
	}
	if ks.IBC.HasKeyExchangeKey {
		t.Errorf("hasKeyExchangeKey 应为 false")
	}
}
