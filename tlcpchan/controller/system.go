package controller

import (
	"net/http"
	"runtime"
	"time"

	"github.com/Trisia/tlcpchan/logger"
)

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemTotalMB   uint64 `json:"mem_total_mb"`
	MemSysMB     uint64 `json:"mem_sys_mb"`
	StartTime    string `json:"start_time"`
	Uptime       string `json:"uptime"`
}

type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

var startTime = time.Now()

type SystemController struct {
	version string
	log     *logger.Logger
}

func NewSystemController(version string) *SystemController {
	return &SystemController{
		version: version,
		log:     logger.Default(),
	}
}

func (c *SystemController) Info(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := SystemInfo{
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemAllocMB:   m.Alloc / 1024 / 1024,
		MemTotalMB:   m.TotalAlloc / 1024 / 1024,
		MemSysMB:     m.Sys / 1024 / 1024,
		StartTime:    startTime.Format(time.RFC3339),
		Uptime:       time.Since(startTime).Round(time.Second).String(),
	}

	Success(w, info)
}

func (c *SystemController) Health(w http.ResponseWriter, r *http.Request) {
	Success(w, HealthStatus{
		Status:  "healthy",
		Version: c.version,
	})
}

func (c *SystemController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/system/info", c.Info)
	router.GET("/api/v1/system/health", c.Health)
}
