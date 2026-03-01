package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Trisia/tlcpchan/config"
)

func TestMCPLogTools(t *testing.T) {
	// 创建临时测试日志文件
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// 写入测试日志内容
	testLogs := `[2025/02/27 10:30:45.123 /test/file.go:100] [DEBUG] 调试消息1
[2025/02/27 10:30:46.234 /test/file.go:101] [INFO] 信息消息1
[2025/02/27 10:30:47.345 /test/file.go:102] [WARN] 警告消息1
[2025/02/27 10:30:48.456 /test/file.go:103] [ERROR] 错误消息1
[2025/02/27 10:30:49.567 /test/file.go:104] [DEBUG] 调试消息2
[2025/02/27 10:30:50.678 /test/file.go:105] [INFO] 信息消息2
[2025/02/27 10:30:51.789 /test/file.go:106] [WARN] 警告消息2
[2025/02/27 10:30:52.890 /test/file.go:107] [ERROR] 错误消息2
`

	if err := os.WriteFile(logFile, []byte(testLogs), 0644); err != nil {
		t.Fatalf("写入测试日志文件失败: %v", err)
	}

	// 创建测试配置
	cfg := &config.Config{
		Server: config.ServerConfig{
			Log: &config.LogConfig{
				File: logFile,
			},
		},
		MCP: config.MCPConfig{
			Enabled: true,
		},
	}

	// 创建 MCPController
	opts := &ServerOptions{
		Config:     cfg,
		ConfigPath: "",
	}
	controller, err := NewMCPController(opts)
	if err != nil {
		t.Fatalf("创建 MCPController 失败: %v", err)
	}

	tests := []struct {
		name          string
		input         GetSystemLogsInput
		expectedCount int
		expectedLevel string
	}{
		{
			name: "读取所有日志",
			input: GetSystemLogsInput{
				Lines: intPtr(10),
			},
			expectedCount: 8,
		},
		{
			name: "读取前3行",
			input: GetSystemLogsInput{
				Lines: intPtr(3),
			},
			expectedCount: 3,
		},
		{
			name: "过滤 INFO 级别",
			input: GetSystemLogsInput{
				Lines: intPtr(10),
				Level: stringPtr("INFO"),
			},
			expectedCount: 2,
			expectedLevel: "INFO",
		},
		{
			name: "过滤 ERROR 级别",
			input: GetSystemLogsInput{
				Lines: intPtr(10),
				Level: stringPtr("ERROR"),
			},
			expectedCount: 2,
			expectedLevel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, output, err := controller.handleGetSystemLogs(context.Background(), nil, tt.input)
			if err != nil {
				t.Fatalf("handleGetSystemLogs 失败: %v", err)
			}

			if len(output.Logs) != tt.expectedCount {
				t.Errorf("期望 %d 条日志，实际得到 %d 条", tt.expectedCount, len(output.Logs))
			}

			if tt.expectedLevel != "" {
				for i, log := range output.Logs {
					if log.Level != tt.expectedLevel {
						t.Errorf("第 %d 条日志级别错误: 期望 %s，实际 %s", i, tt.expectedLevel, log.Level)
					}
				}
			}

			for i, log := range output.Logs {
				if log.Timestamp == "" {
					t.Errorf("第 %d 条日志时间戳为空", i)
				}
				if log.Level == "" {
					t.Errorf("第 %d 条日志级别为空", i)
				}
				if log.Message == "" {
					t.Errorf("第 %d 条日志消息为空", i)
				}
			}
		})
	}
}

func TestTailLines(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	testLogs := `line1
line2
line3
line4
line5
`

	if err := os.WriteFile(logFile, []byte(testLogs), 0644); err != nil {
		t.Fatalf("写入测试日志文件失败: %v", err)
	}

	tests := []struct {
		name     string
		n        int
		filter   string
		expected int
	}{
		{
			name:     "读取最后3行",
			n:        3,
			filter:   "",
			expected: 3,
		},
		{
			name:     "读取所有行",
			n:        10,
			filter:   "",
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := tailLines(logFile, tt.n, tt.filter)
			if err != nil {
				t.Fatalf("tailLines 失败: %v", err)
			}

			if len(lines) != tt.expected {
				t.Errorf("期望 %d 行，实际得到 %d 行", tt.expected, len(lines))
			}
		})
	}
}

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name          string
		line          string
		expectedTS    string
		expectedLevel string
		expectedMsg   string
	}{
		{
			name:          "正常日志行",
			line:          "[2025/02/27 10:30:45.123 /test/file.go:100] [INFO] 测试消息",
			expectedTS:    "2025/02/27",
			expectedLevel: "INFO",
			expectedMsg:   "测试消息",
		},
		{
			name:          "ERROR 级别",
			line:          "[2025/02/27 10:30:45.123 /test/file.go:100] [ERROR] 错误消息",
			expectedTS:    "2025/02/27",
			expectedLevel: "ERROR",
			expectedMsg:   "错误消息",
		},
		{
			name:          "无效格式",
			line:          "invalid log line",
			expectedTS:    "",
			expectedLevel: "",
			expectedMsg:   "invalid log line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := parseLogLine(tt.line)

			if entry.Timestamp != tt.expectedTS {
				t.Errorf("时间戳错误: 期望 %s，实际 %s", tt.expectedTS, entry.Timestamp)
			}
			if entry.Level != tt.expectedLevel {
				t.Errorf("日志级别错误: 期望 %s，实际 %s", tt.expectedLevel, entry.Level)
			}
			if entry.Message != tt.expectedMsg {
				t.Errorf("日志消息错误: 期望 %s，实际 %s", tt.expectedMsg, entry.Message)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
