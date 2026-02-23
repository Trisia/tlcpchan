package controller

import (
	"encoding/json"
	"net/http"

	"github.com/Trisia/tlcpchan/logger"
)

// WriteJSON 写入JSON响应
// 参数:
//   - w: HTTP响应写入器
//   - status: HTTP状态码
//   - data: 响应数据
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.Default().Error("JSON编码失败: %v, 状态码: %d", err, status)
		}
	}
}

// WriteError 写入错误响应
// 参数:
//   - w: HTTP响应写入器
//   - status: HTTP状态码
//   - message: 错误消息
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(message)); err != nil {
		logger.Default().Error("写入错误响应失败: %v, 状态码: %d, 原始消息: %s", err, status, message)
	}
	// 记录错误响应日志
	if status >= 400 {
		logger.Default().Error("返回错误响应, 状态码: %d, 消息: %s", status, message)
	}
}

func BadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message)
}

func NotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message)
}

func InternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message)
}

func Success(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusOK, data)
}

func Created(w http.ResponseWriter, data interface{}) {
	WriteJSON(w, http.StatusCreated, data)
}

// SuccessText 写入成功文本响应
// 参数:
//   - w: HTTP响应写入器
//   - text: 文本内容
func SuccessText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(text)); err != nil {
		logger.Default().Error("写入文本响应失败: %v, 文本: %s", err, text)
	}
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
