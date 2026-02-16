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

type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
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

/**
 * @api {get} /api/v1/system/info 获取系统信息
 * @apiName GetSystemInfo
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统运行时信息
 *
 * @apiSuccess {String} go_version Go版本
 * @apiSuccess {String} os 操作系统
 * @apiSuccess {String} arch 架构
 * @apiSuccess {Number} num_cpu CPU核心数
 * @apiSuccess {Number} num_goroutine Goroutine数量
 * @apiSuccess {Number} mem_alloc_mb 已分配内存（MB）
 * @apiSuccess {Number} mem_total_mb 总分配内存（MB）
 * @apiSuccess {Number} mem_sys_mb 系统内存（MB）
 * @apiSuccess {String} start_time 启动时间
 * @apiSuccess {String} uptime 运行时长
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "go_version": "go1.21.0",
 *       "os": "linux",
 *       "arch": "amd64",
 *       "num_cpu": 8,
 *       "num_goroutine": 25,
 *       "mem_alloc_mb": 10,
 *       "mem_total_mb": 50,
 *       "mem_sys_mb": 100,
 *       "start_time": "2024-01-01T00:00:00Z",
 *       "uptime": "24h0m0s"
 *     }
 */
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

/**
 * @api {get} /api/v1/system/health 系统健康检查
 * @apiName GetSystemHealth
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 检查系统健康状态
 *
 * @apiSuccess {String} status 健康状态
 * @apiSuccess {String} version 版本号
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "healthy",
 *       "version": "1.0.0"
 *     }
 */
func (c *SystemController) Health(w http.ResponseWriter, r *http.Request) {
	Success(w, HealthStatus{
		Status:  "healthy",
		Version: c.version,
	})
}

/**
 * @api {get} /api/v1/system/version 获取版本信息
 * @apiName GetVersion
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统版本信息
 *
 * @apiSuccess {String} version 版本号
 * @apiSuccess {String} go_version Go版本
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "version": "1.0.0",
 *       "go_version": "go1.21.0"
 *     }
 */
func (c *SystemController) Version(w http.ResponseWriter, r *http.Request) {
	Success(w, VersionInfo{
		Version:   c.version,
		GoVersion: runtime.Version(),
	})
}

func (c *SystemController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/system/info", c.Info)
	router.GET("/api/v1/system/health", c.Health)
	router.GET("/api/v1/system/version", c.Version)
}
