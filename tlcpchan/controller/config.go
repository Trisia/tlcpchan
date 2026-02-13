package controller

import (
	"net/http"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
)

type ConfigController struct {
	configPath string
	cfg        *config.Config
	log        *logger.Logger
}

func NewConfigController(cfg *config.Config, configPath string) *ConfigController {
	return &ConfigController{
		configPath: configPath,
		cfg:        cfg,
		log:        logger.Default(),
	}
}

func (c *ConfigController) Get(w http.ResponseWriter, r *http.Request) {
	Success(w, c.cfg)
}

func (c *ConfigController) Update(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := parseJSON(r, &newCfg); err != nil {
		BadRequest(w, "无效的请求体: "+err.Error())
		return
	}

	if err := config.Validate(&newCfg); err != nil {
		BadRequest(w, "配置验证失败: "+err.Error())
		return
	}

	if err := config.Save(c.configPath, &newCfg); err != nil {
		InternalError(w, "保存配置失败: "+err.Error())
		return
	}

	c.cfg = &newCfg
	c.log.Info("配置已更新")
	Success(w, c.cfg)
}

func (c *ConfigController) Reload(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(c.configPath)
	if err != nil {
		InternalError(w, "重新加载配置失败: "+err.Error())
		return
	}

	c.cfg = cfg
	c.log.Info("配置已重新加载")
	Success(w, c.cfg)
}

func (c *ConfigController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/config", c.Get)
	router.POST("/api/v1/config", c.Update)
	router.POST("/api/v1/config/reload", c.Reload)
}
