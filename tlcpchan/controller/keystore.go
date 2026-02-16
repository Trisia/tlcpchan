package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Trisia/tlcpchan/key"
	"github.com/Trisia/tlcpchan/logger"
)

// KeyStoreController 密钥存储控制器
type KeyStoreController struct {
	keyMgr *key.Manager
	log    *logger.Logger
}

// NewKeyStoreController 创建密钥存储控制器
func NewKeyStoreController(keyMgr *key.Manager) *KeyStoreController {
	return &KeyStoreController{
		keyMgr: keyMgr,
		log:    logger.Default(),
	}
}

/**
 * @api {get} /api/v1/keystores 获取密钥列表
 * @apiName ListKeyStores
 * @apiGroup KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取所有密钥存储列表
 *
 * @apiSuccess {Object[]} keystores 密钥列表
 * @apiSuccess {String} keystores.name 密钥名称
 * @apiSuccess {String} keystores.type 类型（tlcp/tls）
 * @apiSuccess {Object} keystores.keyParams 密钥参数
 * @apiSuccess {Boolean} keystores.hasSignCert 是否有签名证书
 * @apiSuccess {Boolean} keystores.hasSignKey 是否有签名密钥
 * @apiSuccess {Boolean} keystores.hasEncCert 是否有加密证书（仅国密）
 * @apiSuccess {Boolean} keystores.hasEncKey 是否有加密密钥（仅国密）
 * @apiSuccess {String} keystores.createdAt 创建时间
 * @apiSuccess {String} keystores.updatedAt 更新时间
 */
func (c *KeyStoreController) List(w http.ResponseWriter, r *http.Request) {
	list, err := c.keyMgr.List()
	if err != nil {
		InternalError(w, "获取密钥列表失败: "+err.Error())
		return
	}
	Success(w, map[string]interface{}{"keystores": list})
}

/**
 * @api {get} /api/v1/keystores/:name 获取密钥详情
 * @apiName GetKeyStore
 * @apiGroup KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取指定密钥的详细信息
 */
func (c *KeyStoreController) Get(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if name == "" {
		BadRequest(w, "缺少密钥名称")
		return
	}

	info, err := c.keyMgr.GetInfo(name)
	if err != nil {
		NotFound(w, "密钥不存在")
		return
	}

	Success(w, info)
}

/**
 * @api {post} /api/v1/keystores 创建密钥
 * @apiName CreateKeyStore
 * @apiGroup KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 创建新的密钥存储（multipart/form-data）
 *
 * @apiParam {String} name 密钥名称
 * @apiParam {String} type 类型（tlcp/tls）
 * @apiParam {String} [keyParams.algorithm] 算法（SM2/RSA/ECDSA）
 * @apiParam {Number} [keyParams.length] 密钥长度
 * @apiParam {File} [signCert] 签名证书文件
 * @apiParam {File} [signKey] 签名密钥文件
 * @apiParam {File} [encCert] 加密证书文件（仅国密）
 * @apiParam {File} [encKey] 加密密钥文件（仅国密）
 */
func (c *KeyStoreController) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		BadRequest(w, "解析上传文件失败: "+err.Error())
		return
	}

	name := r.FormValue("name")
	keyType := key.KeyStoreType(r.FormValue("type"))
	algorithm := r.FormValue("keyParams.algorithm")
	lengthStr := r.FormValue("keyParams.length")
	length := 2048
	if lengthStr != "" {
		_, _ = fmt.Sscanf(lengthStr, "%d", &length)
	}

	if name == "" {
		BadRequest(w, "缺少名称")
		return
	}
	if keyType != key.KeyStoreTypeTLCP && keyType != key.KeyStoreTypeTLS {
		BadRequest(w, "无效的类型，必须是 tlcp 或 tls")
		return
	}

	signCert := readFormFile(r, "signCert")
	signKey := readFormFile(r, "signKey")
	encCert := readFormFile(r, "encCert")
	encKey := readFormFile(r, "encKey")

	keyParams := key.KeyParams{
		Algorithm: algorithm,
		Length:    length,
		Type:      string(keyType),
	}

	ks, err := c.keyMgr.Create(name, keyType, keyParams, signCert, signKey, encCert, encKey)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("创建密钥: %s", name)
	Created(w, ks)
}

/**
 * @api {post} /api/v1/keystores/:name/certificates 更新证书
 * @apiName UpdateKeyStoreCertificates
 * @apiGroup KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 更新密钥的证书（密钥保持不变）
 *
 * @apiParam {File} [signCert] 签名证书文件
 * @apiParam {File} [encCert] 加密证书文件（仅国密）
 */
func (c *KeyStoreController) UpdateCertificates(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if name == "" {
		BadRequest(w, "缺少密钥名称")
		return
	}

	if err := r.ParseMultipartForm(50 << 20); err != nil {
		BadRequest(w, "解析上传文件失败: "+err.Error())
		return
	}

	signCert := readFormFile(r, "signCert")
	encCert := readFormFile(r, "encCert")

	ks, err := c.keyMgr.UpdateCertificates(name, signCert, encCert)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	c.log.Info("更新密钥证书: %s", name)
	Success(w, ks)
}

/**
 * @api {delete} /api/v1/keystores/:name 删除密钥
 * @apiName DeleteKeyStore
 * @apiGroup KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 删除指定的密钥存储
 */
func (c *KeyStoreController) Delete(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if name == "" {
		BadRequest(w, "缺少密钥名称")
		return
	}

	if err := c.keyMgr.Delete(name); err != nil {
		NotFound(w, "密钥不存在")
		return
	}

	c.log.Info("删除密钥: %s", name)
	NoContent(w)
}

/**
 * @api {post} /api/v1/keystores/:name/reload 重载密钥
 * @apiName ReloadKeyStore
 * @apiGroup KeyStore
 * @apiVersion 1.0.0
 *
 * @apiDescription 重载指定密钥
 */
func (c *KeyStoreController) Reload(w http.ResponseWriter, r *http.Request) {
	name := PathParam(r, "name")
	if name == "" {
		BadRequest(w, "缺少密钥名称")
		return
	}

	if !c.keyMgr.Exists(name) {
		NotFound(w, "密钥不存在")
		return
	}

	c.log.Info("重载密钥: %s", name)
	Success(w, map[string]string{"message": "密钥已重载"})
}

// RegisterRoutes 注册路由
func (c *KeyStoreController) RegisterRoutes(router *Router) {
	router.GET("/api/v1/keystores", c.List)
	router.GET("/api/v1/keystores/:name", c.Get)
	router.POST("/api/v1/keystores", c.Create)
	router.POST("/api/v1/keystores/:name/certificates", c.UpdateCertificates)
	router.DELETE("/api/v1/keystores/:name", c.Delete)
	router.POST("/api/v1/keystores/:name/reload", c.Reload)
}

func readFormFile(r *http.Request, fieldName string) []byte {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return nil
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil
	}
	return data
}
