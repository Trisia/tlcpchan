package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	// LevelDebug 调试级别，输出详细调试信息
	LevelDebug Level = iota
	// LevelInfo 信息级别，输出常规运行信息
	LevelInfo
	// LevelWarn 警告级别，输出警告信息
	LevelWarn
	// LevelError 错误级别，输出错误信息
	LevelError
	// LevelFatal 致命级别，输出后退出程序
	LevelFatal
	// LevelDisabled 禁用日志
	LevelDisabled
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// LogConfig 日志配置
type LogConfig struct {
	// Level 日志级别，可选值: "debug", "info", "warn", "error"
	Level string
	// File 日志文件路径，为空则仅输出到控制台
	File string
	// MaxSize 单个日志文件最大大小，单位: MB
	MaxSize int
	// MaxBackups 保留的旧日志文件最大数量，单位: 个
	MaxBackups int
	// MaxAge 保留旧日志文件的最大天数，单位: 天
	MaxAge int
	// Compress 是否压缩旧日志文件
	Compress bool
	// Enabled 是否启用日志
	Enabled bool
}

// Logger 日志记录器，支持文件输出、日志轮转和压缩
type Logger struct {
	*log.Logger
	mu sync.Mutex
	// level 当前日志级别
	level Level
	// enabled 是否启用
	enabled bool
	// writer 输出目标
	writer io.Writer
	// file 日志文件句柄
	file   *os.File
	config LogConfig
	// filePath 日志文件路径
	filePath string
	// curSize 当前日志文件大小，单位: 字节
	curSize int64
}

var (
	defaultLogger *Logger
	once          sync.Once
	mu            sync.Mutex
)

func Default() *Logger {
	once.Do(func() {
		defaultLogger = &Logger{
			Logger:  log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
			level:   LevelInfo,
			enabled: true,
			writer:  os.Stdout,
		}
	})
	return defaultLogger
}

func Init(cfg LogConfig) (*Logger, error) {
	mu.Lock()
	defer mu.Unlock()

	l := &Logger{
		level:   ParseLevel(cfg.Level),
		enabled: cfg.Enabled,
		config:  cfg,
	}

	if !cfg.Enabled {
		l.Logger = log.New(io.Discard, "", 0)
		l.writer = io.Discard
		return l, nil
	}

	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if cfg.File != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}

		f, err := os.OpenFile(cfg.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %w", err)
		}

		fi, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("获取文件信息失败: %w", err)
		}

		l.file = f
		l.filePath = cfg.File
		l.curSize = fi.Size()
		writers = append(writers, f)

		go l.cleanupLoop()
	}

	l.writer = io.MultiWriter(writers...)
	l.Logger = log.New(l.writer, "", log.LstdFlags|log.Lshortfile)

	return l, nil
}

func New(cfg LogConfig) (*Logger, error) {
	return Init(cfg)
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = enabled
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) shouldLog(level Level) bool {
	return l.enabled && level >= l.level
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg := fmt.Sprintf("[%s] %s", level.String(), fmt.Sprintf(format, args...))
	l.Logger.Output(3, msg)

	if l.file != nil && l.config.MaxSize > 0 {
		l.curSize += int64(len(msg) + 30)
		if l.curSize >= int64(l.config.MaxSize)*1024*1024 {
			l.rotate()
		}
	}
}

func (l *Logger) rotate() {
	if l.file == nil {
		return
	}

	oldPath := l.filePath
	l.file.Close()

	ts := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s", oldPath, ts)
	os.Rename(oldPath, backupPath)

	if l.config.Compress {
		go compressFile(backupPath)
	}

	f, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("重新打开日志文件失败: %v", err)
		return
	}
	l.file = f
	l.curSize = 0

	writers := []io.Writer{os.Stdout, f}
	l.writer = io.MultiWriter(writers...)
	l.Logger.SetOutput(l.writer)

	l.cleanOldBackups()
}

func (l *Logger) cleanOldBackups() {
	if l.filePath == "" || (l.config.MaxBackups <= 0 && l.config.MaxAge <= 0) {
		return
	}

	dir := filepath.Dir(l.filePath)
	base := filepath.Base(l.filePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []os.FileInfo
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, base+".") && name != base {
			info, err := entry.Info()
			if err == nil {
				backups = append(backups, info)
			}
		}
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime().After(backups[j].ModTime())
	})

	now := time.Now()
	for i, info := range backups {
		shouldDelete := false

		if l.config.MaxBackups > 0 && i >= l.config.MaxBackups {
			shouldDelete = true
		}

		if l.config.MaxAge > 0 {
			if now.Sub(info.ModTime()) > time.Duration(l.config.MaxAge)*24*time.Hour {
				shouldDelete = true
			}
		}

		if shouldDelete {
			path := filepath.Join(dir, info.Name())
			os.Remove(path)
			if strings.HasSuffix(info.Name(), ".gz") {
				os.Remove(strings.TrimSuffix(path, ".gz"))
			}
		}
	}
}

func (l *Logger) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		l.cleanOldBackups()
		l.mu.Unlock()
	}
}

func compressFile(path string) {
	src, err := os.Open(path)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(path + ".gz")
	if err != nil {
		return
	}
	defer dst.Close()

	gz := gzip.NewWriter(dst)
	defer gz.Close()

	io.Copy(gz, src)
	os.Remove(path)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal 输出致命错误日志并退出程序
// 参数:
//   - format: 格式化字符串
//   - args: 格式化参数
//
// 注意: 调用后程序会以状态码1退出
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// Fatalf 输出致命错误日志并退出程序
// 参数:
//   - format: 格式化字符串
//   - args: 格式化参数
//
// 注意: 调用后程序会以状态码1退出
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

func InitDefault(cfg LogConfig) error {
	l, err := Init(cfg)
	if err != nil {
		return err
	}
	defaultLogger = l
	return nil
}

func SetLevel(level Level) {
	Default().SetLevel(level)
}

func SetEnabled(enabled bool) {
	Default().SetEnabled(enabled)
}

func Debug(format string, args ...interface{}) {
	Default().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	Default().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	Default().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	Default().Error(format, args...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(format string, args ...interface{}) {
	Default().Fatal(format, args...)
}

// Fatalf 输出致命错误日志并退出程序
func Fatalf(format string, args ...interface{}) {
	Default().Fatalf(format, args...)
}

func Close() error {
	return Default().Close()
}
