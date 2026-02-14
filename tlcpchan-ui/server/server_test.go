package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestServer_VersionAPI(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantVersion string
	}{
		{"默认版本", "1.0.0", "1.0.0"},
		{"开发版本", "dev", "dev"},
		{"语义版本", "2.1.3", "2.1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srv := New(tmpDir, "http://localhost:30080", tt.version)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/ui/version", nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("状态码应为 %d, 实际为 %d", http.StatusOK, rec.Code)
			}

			var resp versionResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}

			if resp.Code != 0 {
				t.Errorf("响应码应为 0, 实际为 %d", resp.Code)
			}

			if resp.Data.Version != tt.wantVersion {
				t.Errorf("版本应为 %s, 实际为 %s", tt.wantVersion, resp.Data.Version)
			}

			if resp.Data.GoVersion == "" {
				t.Error("Go版本不应为空")
			}
		})
	}
}

func TestServer_CORS(t *testing.T) {
	tmpDir := t.TempDir()
	srv := New(tmpDir, "http://localhost:8080", "1.0.0")

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/ui/version", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("状态码应为 %d, 实际为 %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("CORS Origin 应为 http://example.com, 实际为 %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestServer_APIShouldProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test":"ok"}`))
	}))
	defer backend.Close()

	tmpDir := t.TempDir()
	srv := New(tmpDir, backend.URL, "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("状态码应为 %d, 实际为 %d", http.StatusOK, rec.Code)
	}
}

func TestServer_StaticFiles(t *testing.T) {
	t.Run("不存在的路径返回index.html", func(t *testing.T) {
		tmpDir := t.TempDir()
		indexContent := `<!DOCTYPE html><html><body>Test</body></html>`
		indexPath := filepath.Join(tmpDir, "index.html")
		if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		srv := New(tmpDir, "http://localhost:8080", "1.0.0")

		req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Body.String() != indexContent {
			t.Errorf("应返回index.html内容, 实际为: %s", rec.Body.String())
		}
	})

	t.Run("存在的文件直接返回", func(t *testing.T) {
		tmpDir := t.TempDir()
		fileContent := `{"test": true}`
		filePath := filepath.Join(tmpDir, "test.json")
		if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		srv := New(tmpDir, "http://localhost:8080", "1.0.0")

		req := httptest.NewRequest(http.MethodGet, "/test.json", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)

		if rec.Body.String() != fileContent {
			t.Errorf("应返回文件内容, 实际为: %s", rec.Body.String())
		}
	})
}

func TestServer_VersionNotProxied(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("UI版本API不应被代理到后端")
	}))
	defer backend.Close()

	tmpDir := t.TempDir()
	srv := New(tmpDir, backend.URL, "test-version")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ui/version", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp versionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.Data.Version != "test-version" {
		t.Errorf("版本应为 test-version, 实际为 %s", resp.Data.Version)
	}
}
