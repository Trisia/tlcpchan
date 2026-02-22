package controller

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Trisia/tlcpchan/config"
	"github.com/Trisia/tlcpchan/logger"
)

type LogsController struct {
	cfg *config.Config
	log *logger.Logger
}

func NewLogsController(cfg *config.Config) *LogsController {
	return &LogsController{
		cfg: cfg,
		log: logger.Default(),
	}
}

/**
 * @api {get} /api/system/logs 获取日志文件列表
 * @apiName ListLogs
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 获取系统所有日志文件列表，包括当前日志和历史归档日志
 *
 * @apiSuccess {Object[]} files 日志文件列表
 * @apiSuccess {String} files.name 文件名，如 "tlcpchan.log" 或 "tlcpchan.log.1"
 * @apiSuccess {String} files.path 文件完整路径
 * @apiSuccess {Number} files.size 文件大小，单位：字节
 * @apiSuccess {String} files.modTime 修改时间，ISO 8601 格式
 * @apiSuccess {Boolean} files.current 是否为当前正在写入的日志文件
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "files": [
 *         {
 *           "name": "tlcpchan.log",
 *           "path": "/etc/tlcpchan/logs/tlcpchan.log",
 *           "size": 102400,
 *           "modTime": "2024-01-01T10:30:00Z",
 *           "current": true
 *         },
 *         {
 *           "name": "tlcpchan.log.20240101-100000.gz",
 *           "path": "/etc/tlcpchan/logs/tlcpchan.log.20240101-100000.gz",
 *           "size": 51200,
 *           "modTime": "2024-01-01T10:00:00Z",
 *           "current": false
 *         }
 *       ]
 *     }
 */
func (c *LogsController) List(w http.ResponseWriter, r *http.Request) {
	logDir := c.getLogDir()
	if logDir == "" {
		BadRequest(w, "日志未配置文件输出")
		return
	}

	files, err := c.listLogFiles(logDir)
	if err != nil {
		InternalError(w, "获取日志文件列表失败: "+err.Error())
		return
	}

	Success(w, map[string]interface{}{
		"files": files,
	})
}

/**
 * @api {get} /api/system/logs/content 读取日志内容
 * @apiName ReadLogContent
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 读取指定日志文件的最后N行内容，支持按日志级别过滤
 *
 * @apiQuery {String} [file=tlcpchan.log] 日志文件名，默认读取当前日志文件
 * @apiQuery {Number} [lines=500] 读取行数，默认500行，最大2000行
 * @apiQuery {String} [level] 日志级别过滤，可选值：debug、info、warn、error
 *
 * @apiSuccess {String} file 文件名
 * @apiSuccess {String[]} lines 日志行数组
 * @apiSuccess {Number} total 文件总行数
 * @apiSuccess {Number} returned 实际返回的行数
 *
 * @apiSuccessExample {json} Success-Response:
 *     HTTP/1.1 200 OK
 *     {
 *       "file": "tlcpchan.log",
 *       "lines": [
 *         "2024-01-01T10:30:00.000Z [INFO] proxy server started on :8443",
 *         "2024-01-01T10:30:01.000Z [DEBUG] connection from 192.168.1.1:12345"
 *       ],
 *       "total": 1000,
 *       "returned": 2
 *     }
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     日志文件不存在: tlcpchan.log
 */
func (c *LogsController) ReadContent(w http.ResponseWriter, r *http.Request) {
	logDir := c.getLogDir()
	if logDir == "" {
		BadRequest(w, "日志未配置文件输出")
		return
	}

	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		fileName = "tlcpchan.log"
	}

	filePath := filepath.Join(logDir, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		NotFound(w, "日志文件不存在: "+fileName)
		return
	}

	linesParam := r.URL.Query().Get("lines")
	lines := 500
	if linesParam != "" {
		if n, err := strconv.Atoi(linesParam); err == nil && n > 0 {
			lines = n
			if lines > 2000 {
				lines = 2000
			}
		}
	}

	levelFilter := strings.ToUpper(r.URL.Query().Get("level"))

	content, total, returned, err := c.readLastLines(filePath, lines, levelFilter)
	if err != nil {
		InternalError(w, "读取日志内容失败: "+err.Error())
		return
	}

	Success(w, map[string]interface{}{
		"file":     fileName,
		"lines":    content,
		"total":    total,
		"returned": returned,
	})
}

/**
 * @api {get} /api/system/logs/download/:filename 下载单个日志文件
 * @apiName DownloadLogFile
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 下载指定的日志文件，原始文件流
 *
 * @apiParam {String} filename 日志文件名（路径参数），如 "tlcpchan.log"
 *
 * @apiSuccessExample {binary} Success-Response:
 *     HTTP/1.1 200 OK
 *     Content-Type: application/octet-stream
 *     Content-Disposition: attachment; filename="tlcpchan.log"
 *
 *     [二进制文件内容]
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 404 Not Found
 *     Content-Type: text/plain
 *
 *     日志文件不存在
 */
func (c *LogsController) Download(w http.ResponseWriter, r *http.Request) {
	filename := PathParam(r, "filename")
	if filename == "" {
		BadRequest(w, "文件名不能为空")
		return
	}

	logDir := c.getLogDir()
	if logDir == "" {
		BadRequest(w, "日志未配置文件输出")
		return
	}

	filePath := filepath.Join(logDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		NotFound(w, "日志文件不存在")
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}

/**
 * @api {get} /api/system/logs/download-all 打包下载所有日志文件
 * @apiName DownloadAllLogs
 * @apiGroup System
 * @apiVersion 1.0.0
 *
 * @apiDescription 将所有日志文件打包成ZIP格式下载
 *
 * @apiSuccessExample {binary} Success-Response:
 *     HTTP/1.1 200 OK
 *     Content-Type: application/zip
 *     Content-Disposition: attachment; filename="tlcpchan-logs-20240101-103045.zip"
 *
 *     [ZIP压缩包]
 *
 * @apiErrorExample {text} Error-Response:
 *     HTTP/1.1 400 Bad Request
 *     Content-Type: text/plain
 *
 *     日志未配置文件输出
 */
func (c *LogsController) DownloadAll(w http.ResponseWriter, r *http.Request) {
	logDir := c.getLogDir()
	if logDir == "" {
		BadRequest(w, "日志未配置文件输出")
		return
	}

	files, err := c.listLogFiles(logDir)
	if err != nil {
		InternalError(w, "获取日志文件列表失败: "+err.Error())
		return
	}

	if len(files) == 0 {
		BadRequest(w, "没有可下载的日志文件")
		return
	}

	zipName := fmt.Sprintf("tlcpchan-logs-%s.zip", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+zipName+"\"")
	w.Header().Set("Content-Type", "application/zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, fileInfo := range files {
		filePath := fileInfo.Path

		f, err := os.Open(filePath)
		if err != nil {
			c.log.Error("打开日志文件失败: %s, %v", filePath, err)
			continue
		}

		zipEntry, err := zipWriter.Create(fileInfo.Name)
		if err != nil {
			f.Close()
			c.log.Error("创建ZIP条目失败: %s, %v", fileInfo.Name, err)
			continue
		}

		if _, err := io.Copy(zipEntry, f); err != nil {
			f.Close()
			c.log.Error("写入ZIP条目失败: %s, %v", fileInfo.Name, err)
			continue
		}

		f.Close()
	}
}

func (c *LogsController) RegisterRoutes(router *Router) {
	router.GET("/api/system/logs", c.List)
	router.GET("/api/system/logs/content", c.ReadContent)
	router.GET("/api/system/logs/download/:filename", c.Download)
	router.GET("/api/system/logs/download-all", c.DownloadAll)
}

func (c *LogsController) getLogDir() string {
	if c.cfg != nil && c.cfg.Server.Log != nil && c.cfg.Server.Log.File != "" {
		return filepath.Dir(c.cfg.Server.Log.File)
	}
	return ""
}

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
	Current bool      `json:"current"`
}

func (c *LogsController) listLogFiles(logDir string) ([]FileInfo, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	currentLogFile := ""
	if c.cfg != nil && c.cfg.Server.Log != nil {
		currentLogFile = filepath.Base(c.cfg.Server.Log.File)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !c.isLogFile(name, currentLogFile) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    name,
			Path:    filepath.Join(logDir, name),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Current: name == currentLogFile,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Current {
			return true
		}
		if files[j].Current {
			return false
		}
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

func (c *LogsController) isLogFile(filename, currentLogFile string) bool {
	if filename == currentLogFile {
		return true
	}

	if strings.HasPrefix(filename, currentLogFile+".") {
		return true
	}

	if strings.HasSuffix(filename, ".gz") {
		base := strings.TrimSuffix(filename, ".gz")
		if strings.HasPrefix(base, currentLogFile+".") {
			return true
		}
	}

	return false
}

func (c *LogsController) readLastLines(filePath string, maxLines int, levelFilter string) ([]string, int, int, error) {
	var reader io.Reader
	f, err := os.Open(filePath)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(f)
		if err != nil {
			return nil, 0, 0, err
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = f
	}

	scanner := bufio.NewScanner(reader)
	var lines []string
	total := 0

	for scanner.Scan() {
		total++
		line := scanner.Text()

		if levelFilter != "" {
			levelMarker := "[" + levelFilter + "]"
			if !strings.Contains(line, levelMarker) {
				continue
			}
		}

		lines = append(lines, line)
		if len(lines) > maxLines {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, 0, err
	}

	returned := len(lines)
	return lines, total, returned, nil
}
