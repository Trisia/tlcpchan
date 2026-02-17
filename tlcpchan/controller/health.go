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
	r.GET("/api/health", c.Health)
	r.GET("/health", c.Health)
}

/**
 * @api {get} /api/health 健康检查
 * @apiName HealthCheck
 * @apiGroup Health
 * @apiVersion 1.0.0
 *
 * @apiDescription 检查系统健康状态，返回系统是否正常运行
 *
 * @apiSuccess {String} status 健康状态，固定值 "ok" 表示系统正常
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "ok"
 *     }
 */
/**
 * @api {get} /health 健康检查（备用路径）
 * @apiName HealthCheckAlt
 * @apiGroup Health
 * @apiVersion 1.0.0
 *
 * @apiDescription 检查系统健康状态的备用路径，功能与 /api/health 完全相同
 *
 * @apiSuccess {String} status 健康状态，固定值 "ok" 表示系统正常
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "status": "ok"
 *     }
 */
func (c *HealthController) Health(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]interface{}{
		"status": "ok",
	})
}
