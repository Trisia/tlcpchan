package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
	"github.com/Trisia/tlcpchan/security/certgen"
	"github.com/Trisia/tlcpchan/security/ibcparams"
)

// GenerateIBCParamsRequest 生成测试 KGC 公共参数请求
type GenerateIBCParamsRequest struct {
	// DistrictName KGC 属地区域名，默认 "tlcpchan.local"
	DistrictName string `json:"districtName,omitempty"`
	// DistrictSerial KGC 序号，与 DistrictName 共同唯一标识 KGC，默认 1
	DistrictSerial int `json:"districtSerial,omitempty"`
	// Years 公共参数有效期，单位：年，默认 10
	Years int `json:"years,omitempty"`
}

/**
 * @api {get} /api/security/ibcparams 列出所有 IBC 信任池参数
 * @apiName ListIBCParams
 * @apiGroup Security-IBCParams
 * @apiVersion 1.1.0
 *
 * @apiDescription 获取 IBC 信任池（KGC 公共参数）列表，客户端与服务端共用同一列表
 *
 * @apiSuccess {Object[]} - KGC 公共参数列表数组
 * @apiSuccess {String} -.filename 文件名，唯一标识符
 * @apiSuccess {String} -.districtName KGC 属地区域名
 * @apiSuccess {Number} -.districtSerial 同一区域下的 KGC 序号
 * @apiSuccess {String} -.notBefore 公共参数生效时间，ISO 8601 格式，零值表示不限
 * @apiSuccess {String} -.notAfter 公共参数失效时间，ISO 8601 格式，零值表示不限
 * @apiSuccess {String} -.issuerIdentity 公共参数颁发者标识
 * @apiSuccess {String} -.signKeyFingerprint 签名主公钥 SM3 指纹（HEX）
 * @apiSuccess {String} -.encKeyFingerprint 加密主公钥 SM3 指纹（HEX）
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     [
 *       {
 *         "filename": "tlcpchan-ibc-kgc.pem",
 *         "districtName": "tlcpchan.local",
 *         "districtSerial": 1,
 *         "notBefore": "2024-01-01T00:00:00Z",
 *         "notAfter": "2034-01-01T00:00:00Z",
 *         "issuerIdentity": "tlcpchan.local",
 *         "signKeyFingerprint": "1a2b3c...",
 *         "encKeyFingerprint": "4d5e6f..."
 *       }
 *     ]
 */
func (c *SecurityController) ListIBCParams(w http.ResponseWriter, r *http.Request) {
	Success(w, c.ibcParamMgr.List())
}

/**
 * @api {get} /api/security/ibcparams/:filename 下载 KGC 公共参数
 * @apiName GetIBCParams
 * @apiGroup Security-IBCParams
 * @apiVersion 1.1.0
 *
 * @apiDescription 下载指定 KGC 公共参数文件
 *
 * @apiParam {String} filename 文件名（路径参数），唯一标识符
 *
 * @apiSuccess {File} - 公共参数文件内容（PEM 格式）
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     IBC 公共参数不存在
 */
func (c *SecurityController) GetIBCParams(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	paramsData, err := c.ibcParamMgr.ReadFile(filename)
	if err != nil {
		NotFound(w, "IBC 公共参数不存在")
		return
	}

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(paramsData)
}

/**
 * @api {delete} /api/security/ibcparams/:filename 删除 KGC 公共参数
 * @apiName DeleteIBCParams
 * @apiGroup Security-IBCParams
 * @apiVersion 1.1.0
 *
 * @apiDescription 删除指定的 KGC 公共参数，删除后需重载相关实例才生效
 *
 * @apiParam {String} filename 文件名（路径参数），唯一标识符
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     删除失败: 具体错误信息
 */
func (c *SecurityController) DeleteIBCParams(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	if err := c.ibcParamMgr.Delete(filename); err != nil {
		InternalError(w, "删除失败: "+err.Error())
		return
	}
	c.log.Info("删除 IBC 公共参数: %s", filename)
	Success(w, nil)
}

/**
 * @api {post} /api/security/ibcparams 添加 KGC 公共参数
 * @apiName AddIBCParams
 * @apiGroup Security-IBCParams
 * @apiVersion 1.1.0
 *
 * @apiDescription 上传并添加 KGC 公共参数到 IBC 信任池。
 * 必须通过可信渠道获取 KGC 公共参数后再入库，绝不直接信任对端下发的参数。
 *
 * @apiBody {String} filename 文件名，表单字段，用于指定保存的文件名（必需）
 * @apiBody {File} params 公共参数文件，表单字段，支持 PEM / DER / HEX / Base64
 *
 * @apiSuccess {String} filename 文件名
 * @apiSuccess {String} districtName KGC 属地区域名
 * @apiSuccess {Number} districtSerial 同一区域下的 KGC 序号
 * @apiSuccess {String} notBefore 公共参数生效时间
 * @apiSuccess {String} notAfter 公共参数失效时间
 * @apiSuccess {String} issuerIdentity 公共参数颁发者标识
 * @apiSuccess {String} signKeyFingerprint 签名主公钥 SM3 指纹（HEX）
 * @apiSuccess {String} encKeyFingerprint 加密主公钥 SM3 指纹（HEX）
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "my-kgc.pem",
 *       "districtName": "example.local",
 *       "districtSerial": 1,
 *       "notBefore": "2024-01-01T00:00:00Z",
 *       "notAfter": "2034-01-01T00:00:00Z",
 *       "issuerIdentity": "example.local",
 *       "signKeyFingerprint": "1a2b3c...",
 *       "encKeyFingerprint": "4d5e6f..."
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     解析表单失败: 具体错误信息
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     文件名不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     添加失败: 具体错误信息
 */
func (c *SecurityController) AddIBCParams(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		BadRequest(w, "解析表单失败: "+err.Error())
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		BadRequest(w, "文件名不能为空")
		return
	}
	filename = normalizeIBCFilename(filename)

	file, _, err := r.FormFile("params")
	if err != nil {
		BadRequest(w, "公共参数文件不能为空: "+err.Error())
		return
	}
	defer file.Close()

	paramsData, err := io.ReadAll(file)
	if err != nil {
		BadRequest(w, "读取公共参数文件失败: "+err.Error())
		return
	}

	param, err := c.ibcParamMgr.Add(filename, paramsData)
	if err != nil {
		InternalError(w, "添加失败: "+err.Error())
		return
	}

	c.log.Info("添加 IBC 公共参数: %s", filename)
	Success(w, param)
}

/**
 * @api {post} /api/security/ibcparams/reload 重载 IBC 信任池
 * @apiName ReloadIBCParams
 * @apiGroup Security-IBCParams
 * @apiVersion 1.1.0
 *
 * @apiDescription 重新扫描信任池目录并重建信任池。
 * 与根证书一致，修改信任池后需重载相关实例才生效。
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     null
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     重载失败: 具体错误信息
 */
func (c *SecurityController) ReloadIBCParams(w http.ResponseWriter, r *http.Request) {
	if err := c.ibcParamMgr.Reload(); err != nil {
		InternalError(w, "重载失败: "+err.Error())
		return
	}
	c.log.Info("重载 IBC 信任池")
	Success(w, nil)
}

/**
 * @api {post} /api/security/ibcparams/generate 生成测试 KGC 公共参数
 * @apiName GenerateIBCParams
 * @apiGroup Security-IBCParams
 * @apiVersion 1.1.0
 *
 * @apiDescription 【仅测试】生成一套 SM9 KGC 主密钥与公共参数，公共参数写入信任池，
 * 主密钥保存在 keystores/ 目录（0600，不提供下载；ibcparams/ 只存放信任的 KGC 公共参数）。
 * 生产环境应由外部 KGC 派生用户私钥后导入。
 *
 * @apiBody {String} [districtName="tlcpchan.local"] KGC 属地区域名
 * @apiBody {Number} [districtSerial=1] KGC 序号
 * @apiBody {Number} [years=10] 公共参数有效期，单位：年
 *
 * @apiSuccess {String} filename 生成的公共参数文件名
 * @apiSuccess {String} masterFile 主密钥文件名（位于 keystores/ 目录）
 * @apiSuccess {Object} param 公共参数元信息，字段同添加接口
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "filename": "kgc-tlcpchan.local-2.pem",
 *       "masterFile": "kgc-tlcpchan.local-2-master.key",
 *       "param": {
 *         "filename": "kgc-tlcpchan.local-2.pem",
 *         "districtName": "tlcpchan.local",
 *         "districtSerial": 2,
 *         "notBefore": "2024-01-01T00:00:00Z",
 *         "notAfter": "2034-01-01T00:00:00Z",
 *         "issuerIdentity": "tlcpchan.local",
 *         "signKeyFingerprint": "1a2b3c...",
 *         "encKeyFingerprint": "4d5e6f..."
 *       }
 *     }
 *
 * @apiParamExample {json} Request-Example:
 *     {
 *       "districtName": "tlcpchan.local",
 *       "districtSerial": 2,
 *       "years": 10
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     KGC 区域名称不能为空
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 500 Internal Server Error
 *     生成公共参数失败: 具体错误信息
 */
func (c *SecurityController) GenerateIBCParams(w http.ResponseWriter, r *http.Request) {
	var req GenerateIBCParamsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		BadRequest(w, "无效的请求: "+err.Error())
		return
	}

	if req.DistrictName == "" {
		req.DistrictName = "tlcpchan.local"
	}
	if req.DistrictSerial <= 0 {
		req.DistrictSerial = 1
	}
	if req.Years <= 0 {
		req.Years = 10
	}

	generated, err := certgen.GenerateIBCParams(req.DistrictName, req.DistrictSerial, tlcp.ValidityPeriod{
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(req.Years, 0, 0),
	})
	if err != nil {
		InternalError(w, "生成公共参数失败: "+err.Error())
		return
	}

	filename := fmt.Sprintf("kgc-%s-%d.pem", req.DistrictName, req.DistrictSerial)
	param, err := c.ibcParamMgr.Add(filename, generated.ParamsPEM)
	if err != nil {
		InternalError(w, "写入信任池失败: "+err.Error())
		return
	}

	masterFile := fmt.Sprintf("kgc-%s-%d-master.key", req.DistrictName, req.DistrictSerial)
	masterPath := filepath.Join(c.cfg.GetKeyStoreStoreDir(), masterFile)
	if err := certgen.SaveIBCMasterToFile(generated.MasterPEM, masterPath); err != nil {
		InternalError(w, "保存主密钥失败: "+err.Error())
		return
	}

	c.log.Info("生成测试 KGC 公共参数: %s", filename)
	Success(w, map[string]interface{}{
		"filename":   filename,
		"masterFile": masterFile,
		"param":      param,
	})
}

// loadBuiltinIBCMaster 读取初始化内置的测试 KGC 主密钥。
//
// 返回值：
//   - *certgen.IBCMaster: 内置测试 KGC 的主密钥对
//   - error: 主密钥文件不存在或解析失败时返回错误
//
// 注意事项：
//   - 生成 IBC 身份仅使用内置测试 KGC，生产环境应由外部 KGC 派生后导入
func (c *SecurityController) loadBuiltinIBCMaster() (*certgen.IBCMaster, error) {
	masterPath := filepath.Join(c.cfg.GetKeyStoreStoreDir(), "tlcpchan-ibc-kgc-master.key")
	data, err := os.ReadFile(masterPath)
	if err != nil {
		return nil, fmt.Errorf("读取内置测试 KGC 主密钥失败（请确认已完成初始化）: %w", err)
	}
	master, err := certgen.LoadIBCMasterFromPEM(data)
	if err != nil {
		return nil, err
	}
	return master, nil
}

// normalizeIBCFilename 规范化上传的公共参数文件名：去除路径并统一为白名单扩展名。
//
// 参数：
//   - filename: 上传时提供的文件名
//
// 返回值：
//   - string: 不含目录的文件名；扩展名不在信任池白名单内时替换为 .pem
func normalizeIBCFilename(filename string) string {
	name := filepath.Base(strings.TrimSpace(filename))
	ext := strings.ToLower(filepath.Ext(name))
	for _, allowed := range ibcparams.ParamExtensions {
		if ext == allowed {
			return name
		}
	}
	return strings.TrimSuffix(name, filepath.Ext(name)) + ".pem"
}
