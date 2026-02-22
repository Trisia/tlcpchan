package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/Trisia/tlcpchan/version"
)

func TestSystemController_Version(t *testing.T) {
	ctrl := NewSystemController()
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

	if resp.Version != version.Version {
		t.Errorf("版本应为 %s, 实际为 %s", version.Version, resp.Version)
	}

	if resp.GoVersion == "" {
		t.Error("Go版本不应为空")
	}

	if resp.GoVersion != runtime.Version() {
		t.Errorf("Go版本应为 %s, 实际为 %s", runtime.Version(), resp.GoVersion)
	}
}

func TestSystemController_Health(t *testing.T) {
	ctrl := NewSystemController()
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

	if resp.Version != version.Version {
		t.Errorf("版本应为 %s, 实际为 %s", version.Version, resp.Version)
	}
}

func TestSystemController_Info(t *testing.T) {
	ctrl := NewSystemController()
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
	ctrl := NewSystemController()
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
