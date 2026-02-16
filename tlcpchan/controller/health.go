package controller

import (
	"net/http"

	"github.com/Trisia/tlcpchan/instance"
)

// HealthController 健康检查控制器
type HealthController struct {
	instanceManager *instance.Manager
}

// NewHealthController 创建健康检查控制器
func NewHealthController(instanceManager *instance.Manager) *HealthController {
	return &HealthController{
		instanceManager: instanceManager,
	}
}

// RegisterRoutes 注册路由
func (c *HealthController) RegisterRoutes(r *Router) {
	r.GET("/api/v1/health", c.Health)
	r.GET("/health", c.Health)
}

// Health 健康检查
func (c *HealthController) Health(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]interface{}{
		"status": "ok",
	})
}
