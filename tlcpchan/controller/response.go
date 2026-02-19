package controller

import (
	"encoding/json"
	"net/http"
)

// WriteJSON 写入JSON响应
// 参数:
//   - w: HTTP响应写入器
//   - status: HTTP状态码
//   - data: 响应数据
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError 写入错误响应
// 参数:
//   - w: HTTP响应写入器
//   - status: HTTP状态码
//   - message: 错误消息
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	w.Write([]byte(message))
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

func SuccessText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(text))
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
