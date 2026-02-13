package controller

import (
	"net/http"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/instance"
	"github.com/Trisia/tlcpchan/logger"
)

type InstanceController struct {
	manager *instance.Manager
	log     *logger.Logger
}

func NewInstanceController(mgr *instance.Manager) *InstanceController {
	return &InstanceController{
		manager: mgr,
		log:     logger.Default(),
	}
}

func (c *InstanceController) List(w http.ResponseWriter, r *http.Request) {
	instances := c.manager.List()
	data := make([]map[string]interface{}, len(instances))
	for i, inst := range instances {
		data[i] = map[string]interface{}{
			"name":    inst.Name(),
			"status":  inst.Status(),
			"config":  inst.Config(),
			"enabled": inst.Config().Enabled,
		}
	}
	Success(w, data)
}

func (c *InstanceController) Get(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}
	Success(w, map[string]interface{}{
		"name":   inst.Name(),
		"status": inst.Status(),
		"config": inst.Config(),
	})
}

func (c *InstanceController) Create(w http.ResponseWriter, r *http.Request) {
	var cfg config.InstanceConfig
	if err := parseJSON(r, &cfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	inst, err := c.manager.Create(&cfg)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("创建实例: %s", cfg.Name)
	Created(w, map[string]interface{}{
		"name":   inst.Name(),
		"status": inst.Status(),
	})
}

func (c *InstanceController) Update(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	var cfg config.InstanceConfig
	if err := parseJSON(r, &cfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	cfg.Name = name
	if inst.Status() == instance.StatusRunning {
		if err := inst.Reload(&cfg); err != nil {
			BadRequest(w, err.Error())
			return
		}
	} else {
		NotFound(w, "实例未运行，无法热更新")
		return
	}

	c.log.Info("更新实例配置: %s", name)
	Success(w, map[string]interface{}{
		"name":   inst.Name(),
		"status": inst.Status(),
	})
}

func (c *InstanceController) Delete(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if err := c.manager.Delete(name); err != nil {
		BadRequest(w, err.Error())
		return
	}
	c.log.Info("删除实例: %s", name)
	Success(w, nil)
}

func (c *InstanceController) Start(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	if err := inst.Start(); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("启动实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

func (c *InstanceController) Stop(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	if err := inst.Stop(); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("停止实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

func (c *InstanceController) Reload(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	cfg := inst.Config()
	if err := inst.Reload(cfg); err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("重载实例: %s", name)
	Success(w, map[string]string{"status": string(inst.Status())})
}

func (c *InstanceController) Stats(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	inst, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	Success(w, inst.Stats())
}

func (c *InstanceController) Logs(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	_, ok := c.manager.Get(name)
	if !ok {
		NotFound(w, "实例不存在")
		return
	}

	Success(w, []map[string]interface{}{
		{"timestamp": "2024-01-01T00:00:00Z", "level": "info", "message": "示例日志"},
	})
}

func (c *InstanceController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/instances", c.List)
	router.POST("/api/v1/instances", c.Create)
	router.GET("/api/v1/instances/:name", c.Get)
	router.PUT("/api/v1/instances/:name", c.Update)
	router.DELETE("/api/v1/instances/:name", c.Delete)
	router.POST("/api/v1/instances/:name/start", c.Start)
	router.POST("/api/v1/instances/:name/stop", c.Stop)
	router.POST("/api/v1/instances/:name/reload", c.Reload)
	router.GET("/api/v1/instances/:name/stats", c.Stats)
	router.GET("/api/v1/instances/:name/logs", c.Logs)
}
