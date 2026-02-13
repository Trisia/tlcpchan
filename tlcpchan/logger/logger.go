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

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
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
	default:
		return LevelInfo
	}
}

type LogConfig struct {
	Level      string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Enabled    bool
}

type Logger struct {
	*log.Logger
	mu       sync.Mutex
	level    Level
	enabled  bool
	writer   io.Writer
	file     *os.File
	config   LogConfig
	filePath string
	curSize  int64
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

func Close() error {
	return Default().Close()
}
