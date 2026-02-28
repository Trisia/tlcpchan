package controller

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LogEntry 日志条目
type LogEntry struct {
	// Timestamp 时间戳，格式: "2025/02/27 10:30:45.123"
	Timestamp string `json:"timestamp"`
	// Level 日志级别，可选值: DEBUG、INFO、WARN、ERROR、FATAL
	Level string `json:"level"`
	// Message 日志消息内容
	Message string `json:"message"`
}

// GetSystemLogsInput 获取系统日志输入
type GetSystemLogsInput struct {
	// Lines 读取行数，默认 100，最大 1000
	Lines *int `json:"lines,omitempty"`
	// Level 日志级别过滤，可选值: DEBUG、INFO、WARN、ERROR、FATAL
	Level *string `json:"level,omitempty"`
}

// GetSystemLogsOutput 获取系统日志输出
type GetSystemLogsOutput struct {
	// Logs 日志条目列表
	Logs []LogEntry `json:"logs"`
}

// handleGetSystemLogs 处理获取系统日志请求
//
// 参数:
//   - ctx: 上下文
//   - req: MCP 工具调用请求
//   - input: 获取系统日志输入参数
//
// 返回:
//   - *mcp.CallToolResult: MCP 工具调用结果（可以为 nil，SDK 自动处理）
//   - GetSystemLogsOutput: 获取系统日志输出参数，包含日志条目列表
//   - error: 处理失败时返回时返回错误
//
// 注意:
//   - 从配置中读取日志文件路径
//   - 支持按行数和日志级别过滤
//   - 解析日志行并结构化返回
func (c *MCPController) handleGetSystemLogs(_ context.Context, _ *mcp.CallToolRequest, input GetSystemLogsInput) (
	*mcp.CallToolResult,
	GetSystemLogsOutput,
	error,
) {
	if c.config.Server.Log == nil || c.config.Server.Log.File == "" {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: "日志文件未配置",
				},
			},
			IsError: true,
		}
		return result, GetSystemLogsOutput{}, nil
	}

	logFile := c.config.Server.Log.File

	// 设置默认行数
	lines := 100
	if input.Lines != nil {
		lines = *input.Lines
	}
	if lines < 1 {
		lines = 100
	}
	if lines > 1000 {
		lines = 1000
	}

	// 检查文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return nil, GetSystemLogsOutput{Logs: []LogEntry{}}, nil
	}

	// 读取日志行
	var levelFilter string
	if input.Level != nil {
		levelFilter = strings.ToUpper(*input.Level)
	}

	linesText, err := tailLines(logFile, lines, levelFilter)
	if err != nil {
		result := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("读取日志文件失败: %v", err),
				},
			},
			IsError: true,
		}
		return result, GetSystemLogsOutput{}, nil
	}

	// 解析日志行
	logs := make([]LogEntry, 0, len(linesText))
	for _, line := range linesText {
		entry := parseLogLine(line)
		logs = append(logs, entry)
	}

	return nil, GetSystemLogsOutput{Logs: logs}, nil
}

// tailLines 从文件末尾读取指定行数
//
// 参数:
//   - filePath: 文件路径
//   - n: 读取行数
//   - levelFilter: 日志级别过滤器，为空表示不过滤
//
// 返回:
//   - []string: 日志行数组
//   - error: 读取失败时返回错误
//
// 注意:
//   - 支持 .gz 压缩文件
//   - 自动过滤不符合日志级别的行
func tailLines(filePath string, n int, levelFilter string) ([]string, error) {
	var reader io.Reader
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if strings.HasSuffix(filePath, ".gz") {
		gzReader, err := gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = f
	}

	scanner := bufio.NewScanner(reader)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()

		if levelFilter != "" {
			levelMarker := "[" + levelFilter + "]"
			if !strings.Contains(line, levelMarker) {
				continue
			}
		}

		lines = append(lines, line)
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// parseLogLine 解析日志行
//
// 参数:
//   - line: 日志行文本
//
// 返回:
//   - LogEntry: 解析后的日志条目
//
// 注意:
//   - 日志格式: [2025/02/27 10:30:45.123 /path/to/file:123] [INFO] 日志消息
//   - 提取时间戳、日志级别和消息
func parseLogLine(line string) LogEntry {
	entry := LogEntry{
		Timestamp: "",
		Level:     "",
		Message:   line,
	}

	re := regexp.MustCompile(`^\[([^\]]+)\] \[([^\]]+)\] (.+)$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) == 4 {
		timeAndFile := matches[1]
		spaceIndex := strings.Index(timeAndFile, " ")
		if spaceIndex > 0 {
			entry.Timestamp = timeAndFile[:spaceIndex]
		} else {
			entry.Timestamp = timeAndFile
		}

		entry.Level = matches[2]
		entry.Message = matches[3]
	}

	return entry
}

// registerLogTools 注册日志管理工具到 MCP 服务器
//
// 注意:
//   - 注册 get_system_logs 工具
//   - 工具由 MCPController 的 handleGetSystemLogs 处理
func (c *MCPController) registerLogTools() {
	mcp.AddTool(c.server, &mcp.Tool{
		Name:        "get_system_logs",
		Description: "获取系统日志（历史日志文件）",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"lines": map[string]any{
					"description": "读取行数，默认 100，最大 1000",
					"type":        "integer",
					"minimum":     1,
					"maximum":     1000,
				},
				"level": map[string]any{
					"description": "日志级别过滤，可选值: DEBUG、INFO、WARN、ERROR、FATAL",
					"type":        "string",
					"enum":        []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
				},
			},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"logs": map[string]any{
					"description": "日志条目列表",
					"type":        "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"timestamp": map[string]any{
								"description": "时间戳",
								"type":        "string",
							},
							"level": map[string]any{
								"description": "日志级别",
								"type":        "string",
							},
							"message": map[string]any{
								"description": "日志消息",
								"type":        "string",
							},
						},
					},
				},
			},
		},
	}, c.handleGetSystemLogs)
}
