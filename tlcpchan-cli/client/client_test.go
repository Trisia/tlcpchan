package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		goVersion   string
		wantVersion string
	}{
		{"默认版本", "1.0.0", "go1.21.0", "1.0.0"},
		{"开发版本", "dev", "go1.22.0", "dev"},
		{"语义版本", "2.1.3", "go1.23.0", "2.1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/system/version" {
					t.Errorf("请求路径应为 /api/v1/system/version, 实际为 %s", r.URL.Path)
				}

				resp := map[string]string{
					"version":    tt.version,
					"go_version": tt.goVersion,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			info, err := client.GetVersion()
			if err != nil {
				t.Fatalf("GetVersion() 失败: %v", err)
			}

			if info.Version != tt.wantVersion {
				t.Errorf("版本应为 %s, 实际为 %s", tt.wantVersion, info.Version)
			}

			if info.GoVersion != tt.goVersion {
				t.Errorf("Go版本应为 %s, 实际为 %s", tt.goVersion, info.GoVersion)
			}
		})
	}
}

func TestClient_GetVersion_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetVersion()
	if err == nil {
		t.Error("期望返回错误，但返回了 nil")
	}
}

func TestClient_GetVersion_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetVersion()
	if err == nil {
		t.Error("期望返回错误，但返回了 nil")
	}
}

func TestClient_HealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/system/health" {
			t.Errorf("请求路径应为 /api/v1/system/health, 实际为 %s", r.URL.Path)
		}

		resp := map[string]string{
			"status":  "healthy",
			"version": "1.0.0",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	health, err := client.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck() 失败: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("状态应为 healthy, 实际为 %s", health.Status)
	}

	if health.Version != "1.0.0" {
		t.Errorf("版本应为 1.0.0, 实际为 %s", health.Version)
	}
}

func TestClient_ConnectionError(t *testing.T) {
	client := NewClient("http://localhost:99999")
	_, err := client.GetVersion()
	if err == nil {
		t.Error("期望返回连接错误，但返回了 nil")
	}
}
