package controller

import (
	"net/http"
	"runtime"
	"time"

	"github.com/Trisia/tlcpchan/logger"
	"github.com/Trisia/tlcpchan/version"
)

type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"numCpu"`
	NumGoroutine int    `json:"numGoroutine"`
	MemAllocMB   uint64 `json:"memAllocMb"`
	MemTotalMB   uint64 `json:"memTotalMb"`
	MemSysMB     uint64 `json:"memSysMb"`
	StartTime    string `json:"startTime"`
	Uptime       string `json:"uptime"`
}

type HealthStatus struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type VersionInfo struct {
	Version string `json:"version"`
}

var startTime = time.Now()

type SystemController struct {
	log *logger.Logger
}

func NewSystemController() *SystemController {
	return &SystemController{
		log: logger.Default(),
	}
}

/**
 * @api {get} /api/system/info 获取系统信息
 * @apiName GetSystemInfo
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统运行时信息
 *
 * @apiSuccess {String} os 操作系统
 * @apiSuccess {String} arch 架构
 * @apiSuccess {Number} numCpu CPU核心数
 * @apiSuccess {Number} numGoroutine Goroutine数量
 * @apiSuccess {Number} memAllocMb 已分配内存（MB）
 * @apiSuccess {Number} memTotalMb 总分配内存（MB）
 * @apiSuccess {Number}) memSysMb 系统内存（MB）
 * @apiSuccess {String} startTime 启动时间
 * @apiSuccess {String} uptime 运行时长
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "os": "linux",
 *       "arch": "amd64",
 *       "numCpu": 8,
 *       "numGoroutine": 25,
 *       "memAllocMb": 10,
 *       "memTotalMb": 50,
 *       "memSysMb": 100,
 *       "startTime": "2024-01-01T00:00:00Z",
 *       "uptime": "24h0m0s"
 *     }
 */
func (c *SystemController) Info(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := SystemInfo{
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
 * @api {get} /api/system/health 系统健康检查
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
		Version: version.Version,
	})
}

/**
 * @api {get} /api/system/version 获取版本信息
 * @apiName GetVersion
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统版本信息
 *
 * @apiSuccess {String} version 版本号
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "version": "1.0.0"
 *     }
 */
func (c *SystemController) Version(w http.ResponseWriter, r *http.Request) {
	Success(w, VersionInfo{
		Version: version.Version,
	})
}

func (c *SystemController) RegisterRoutes(router *Router) {
	router.GET("/api/system/info", c.Info)
	router.GET("/api/system/health", c.Health)
	router.GET("/api/system/version", c.Version)
	router.GET("/api/version", c.Version)
	router.GET("/version", c.Version)
}
