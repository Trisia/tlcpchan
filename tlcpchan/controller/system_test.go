package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
)

func TestSystemController_Version(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantVersion string
	}{
		{"默认版本", "1.0.0", "1.0.0"},
		{"开发版本", "dev", "dev"},
		{"语义版本", "2.1.3", "2.1.3"},
		{"带前缀版本", "v1.2.3", "v1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewSystemController(tt.version)
			router := NewRouter()
			ctrl.RegisterRoutes(router)

			req := httptest.NewRequest(http.MethodGet, "/api/system/version", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("状态码应为 %d, 实际为 %d", http.StatusOK, rec.Code)
			}

			var resp VersionInfo
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}

			if resp.Version != tt.wantVersion {
				t.Errorf("版本应为 %s, 实际为 %s", tt.wantVersion, resp.Version)
			}

			if resp.GoVersion == "" {
				t.Error("Go版本不应为空")
			}

			if resp.GoVersion != runtime.Version() {
				t.Errorf("Go版本应为 %s, 实际为 %s", runtime.Version(), resp.GoVersion)
			}
		})
	}
}

func TestSystemController_Health(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"健康检查-默认版本", "1.0.0"},
		{"健康检查-开发版本", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewSystemController(tt.version)
			router := NewRouter()
			ctrl.RegisterRoutes(router)

			req := httptest.NewRequest(http.MethodGet, "/api/system/health", nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("状态码应为 %d, 实际为 %d", http.StatusOK, rec.Code)
			}

			var resp HealthStatus
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}

			if resp.Status != "healthy" {
				t.Errorf("状态应为 healthy, 实际为 %s", resp.Status)
			}

			if resp.Version != tt.version {
				t.Errorf("版本应为 %s, 实际为 %s", tt.version, resp.Version)
			}
		})
	}
}

func TestSystemController_Info(t *testing.T) {
	ctrl := NewSystemController("1.0.0")
	router := NewRouter()
	ctrl.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/api/system/info", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("状态码应为 %d, 实际为 %d", http.StatusOK, rec.Code)
	}

	var resp SystemInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if resp.GoVersion == "" {
		t.Error("Go版本不应为空")
	}

	if resp.OS == "" {
		t.Error("操作系统不应为空")
	}

	if resp.Arch == "" {
		t.Error("架构不应为空")
	}

	if resp.NumCPU <= 0 {
		t.Errorf("CPU核心数应大于0, 实际为 %d", resp.NumCPU)
	}
}

func TestSystemController_Routes(t *testing.T) {
	ctrl := NewSystemController("1.0.0")
	router := NewRouter()
	ctrl.RegisterRoutes(router)

	routes := []struct {
		method  string
		pattern string
	}{
		{http.MethodGet, "/api/system/info"},
		{http.MethodGet, "/api/system/health"},
		{http.MethodGet, "/api/system/version"},
	}

	for _, route := range routes {
		found := false
		for _, r := range router.routes {
			if r.Method == route.method && r.Pattern == route.pattern {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("路由 %s %s 未注册", route.method, route.pattern)
		}
	}
}
