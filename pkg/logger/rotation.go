package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// RotationWriter 日志分割写入器
type RotationWriter struct {
	config      FileConfig
	rotation    RotationConfig
	currentFile string
	lumberjack  *lumberjack.Logger
	timeRotate  *TimeRotationWriter
	mu          sync.Mutex
}

// TimeRotationWriter 时间分割写入器
type TimeRotationWriter struct {
	config      TimeRotationConfig
	filename    string
	currentFile string
	lastRotate  time.Time
	mu          sync.Mutex
	// 性能优化：文件句柄缓存和批量写入
	currentFileHandle *os.File
	buffer            []byte
	bufferSize        int
	lastFlush         time.Time
	flushInterval     time.Duration
	// 字符串构建缓存
	filenameBuilder strings.Builder
}

// NewTimeRotationWriter 创建时间分割写入器
func NewTimeRotationWriter(filename string, config TimeRotationConfig) *TimeRotationWriter {
	// 设置默认缓冲区大小和刷新间隔
	bufferSize := 64 * 1024          // 64KB 默认缓冲区
	flushInterval := 5 * time.Second // 5秒刷新间隔

	tr := &TimeRotationWriter{
		config:        config,
		filename:      filename,
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		buffer:        make([]byte, 0, bufferSize),
		lastFlush:     time.Now(),
	}
	tr.updateCurrentFile()
	return tr
}

// NewRotationWriter 创建新的分割写入器
func NewRotationWriter(fileConfig FileConfig, rotationConfig RotationConfig) (*RotationWriter, error) {
	// 确保目录存在
	dir := filepath.Dir(fileConfig.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	rw := &RotationWriter{
		config:      fileConfig,
		rotation:    rotationConfig,
		currentFile: fileConfig.Filename,
	}

	// 初始化大小分割
	if rotationConfig.Size.Enabled {
		rw.lumberjack = &lumberjack.Logger{
			Filename:   fileConfig.Filename,
			MaxSize:    rotationConfig.Size.MaxSize,
			MaxBackups: fileConfig.MaxBackups,
			MaxAge:     fileConfig.MaxAge,
			Compress:   fileConfig.Compress,
			LocalTime:  fileConfig.LocalTime,
		}
	}

	// 初始化时间分割
	if rotationConfig.Time.Enabled {
		rw.timeRotate = &TimeRotationWriter{
			config:   rotationConfig.Time,
			filename: fileConfig.Filename,
		}
		rw.timeRotate.updateCurrentFile()
	}

	return rw, nil
}

// Write 实现io.Writer接口
func (rw *RotationWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// 时间分割优先
	if rw.rotation.Time.Enabled && rw.timeRotate != nil {
		return rw.timeRotate.Write(p)
	}

	// 大小分割
	if rw.rotation.Size.Enabled && rw.lumberjack != nil {
		return rw.lumberjack.Write(p)
	}

	// 默认写入文件
	return rw.writeToFile(p)
}

// writeToFile 写入文件
func (rw *RotationWriter) writeToFile(p []byte) (n int, err error) {
	file, err := os.OpenFile(rw.currentFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return file.Write(p)
}

// Sync 同步数据
func (rw *RotationWriter) Sync() error {
	if rw.lumberjack != nil {
		// lumberjack没有Sync方法，这里不需要处理
	}
	return nil
}

// Close 关闭写入器
func (rw *RotationWriter) Close() error {
	if rw.lumberjack != nil {
		return rw.lumberjack.Close()
	}
	if rw.timeRotate != nil {
		return rw.timeRotate.Close()
	}
	return nil
}

// Close 关闭写入器（优化版本：确保缓冲区刷新和文件句柄关闭）
func (tr *TimeRotationWriter) Close() error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// 刷新缓冲区
	if err := tr.flushBuffer(); err != nil {
		return err
	}

	// 关闭文件句柄
	if tr.currentFileHandle != nil {
		if err := tr.currentFileHandle.Close(); err != nil {
			return err
		}
		tr.currentFileHandle = nil
	}

	return nil
}

// Rotate 手动触发分割
func (rw *RotationWriter) Rotate() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.lumberjack != nil {
		return rw.lumberjack.Rotate()
	}

	if rw.timeRotate != nil {
		return rw.timeRotate.rotate()
	}

	return nil
}

// Write 时间分割写入（优化版本：减少文件操作，使用缓冲写入）
func (tr *TimeRotationWriter) Write(p []byte) (n int, err error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	// 检查是否需要分割
	if tr.shouldRotate() {
		if err = tr.flushBuffer(); err != nil {
			return 0, err
		}
		if err = tr.rotate(); err != nil {
			return 0, err
		}
	}

	// 确保文件句柄可用
	if err = tr.ensureFileHandle(); err != nil {
		return 0, err
	}

	// 使用缓冲写入
	if tr.bufferSize > 0 {
		return tr.writeToBuffer(p)
	}

	// 直接写入文件
	return tr.currentFileHandle.Write(p)
}

// shouldRotate 检查是否应该分割
func (tr *TimeRotationWriter) shouldRotate() bool {
	now := time.Now()

	switch tr.config.Interval {
	case "hour":
		return now.Hour() != tr.lastRotate.Hour() || now.Day() != tr.lastRotate.Day()
	case "day":
		return now.Day() != tr.lastRotate.Day() || now.Month() != tr.lastRotate.Month()
	case "week":
		_, thisWeek := now.ISOWeek()
		_, lastWeek := tr.lastRotate.ISOWeek()
		return thisWeek != lastWeek
	case "month":
		return now.Month() != tr.lastRotate.Month() || now.Year() != tr.lastRotate.Year()
	default:
		return false
	}
}

// rotate 执行分割
func (tr *TimeRotationWriter) rotate() error {
	// 更新当前文件名
	tr.updateCurrentFile()
	tr.lastRotate = time.Now()
	return nil
}

// ensureFileHandle 确保文件句柄可用
func (tr *TimeRotationWriter) ensureFileHandle() error {
	if tr.currentFileHandle == nil || tr.currentFileHandle.Name() != tr.currentFile {
		// 关闭旧文件句柄
		if tr.currentFileHandle != nil {
			tr.currentFileHandle.Close()
		}

		// 确保目录存在
		dir := filepath.Dir(tr.currentFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// 打开新文件句柄
		file, err := os.OpenFile(tr.currentFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		tr.currentFileHandle = file
	}
	return nil
}

// writeToBuffer 写入缓冲区
func (tr *TimeRotationWriter) writeToBuffer(p []byte) (n int, err error) {
	// 检查缓冲区容量
	if len(tr.buffer)+len(p) > tr.bufferSize {
		// 刷新缓冲区
		if err = tr.flushBuffer(); err != nil {
			return 0, err
		}
	}

	// 添加到缓冲区
	tr.buffer = append(tr.buffer, p...)

	// 检查是否需要定时刷新
	if time.Since(tr.lastFlush) > tr.flushInterval {
		if err = tr.flushBuffer(); err != nil {
			return len(p), err // 返回写入长度，即使刷新失败
		}
	}

	return len(p), nil
}

// flushBuffer 刷新缓冲区
func (tr *TimeRotationWriter) flushBuffer() error {
	if len(tr.buffer) == 0 {
		return nil
	}

	if tr.currentFileHandle == nil {
		return fmt.Errorf("file handle is nil")
	}

	_, err := tr.currentFileHandle.Write(tr.buffer)
	if err != nil {
		return err
	}

	// 同步到磁盘
	if err = tr.currentFileHandle.Sync(); err != nil {
		return err
	}

	// 清空缓冲区
	tr.buffer = tr.buffer[:0]
	tr.lastFlush = time.Now()

	return nil
}

// updateCurrentFile 更新当前文件名（优化版本：减少字符串分配）
func (tr *TimeRotationWriter) updateCurrentFile() {
	now := time.Now()
	ext := filepath.Ext(tr.filename)
	base := strings.TrimSuffix(tr.filename, ext)

	// 重置字符串构建器
	tr.filenameBuilder.Reset()
	tr.filenameBuilder.Grow(len(base) + 20) // 预分配容量
	tr.filenameBuilder.WriteString(base)

	switch tr.config.Interval {
	case "hour":
		tr.filenameBuilder.WriteString(now.Format(".2006-01-02-15"))
	case "day":
		tr.filenameBuilder.WriteString(now.Format(".2006-01-02"))
	case "week":
		year, week := now.ISOWeek()
		tr.filenameBuilder.WriteString(fmt.Sprintf(".%d-W%02d", year, week))
	case "month":
		tr.filenameBuilder.WriteString(now.Format(".2006-01"))
	default:
		tr.filenameBuilder.WriteString(now.Format(".2006-01-02-15-04-05"))
	}

	tr.filenameBuilder.WriteString(ext)
	tr.currentFile = tr.filenameBuilder.String()
}

// CreateRotationCore 创建支持分割的zapcore.Core
func CreateRotationCore(config *Config) (zapcore.Core, error) {
	var cores []zapcore.Core

	// 控制台输出
	if config.Output.Console.Enabled {
		consoleEncoder := zapcore.NewConsoleEncoder(buildConsoleEncoderConfig(config))
		consoleWriter := zapcore.Lock(os.Stdout)
		consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, ParseLevel(config.Level))
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if config.Output.File.Enabled {
		fileCore, err := createFileCore(config.Output.File, config.Rotation, config)
		if err != nil {
			return nil, err
		}
		cores = append(cores, fileCore)
	}

	// 多文件输出
	for _, fileConfig := range config.Output.MultiFile {
		if fileConfig.Enabled {
			fileCore, err := createFileCore(fileConfig, config.Rotation, config)
			if err != nil {
				return nil, err
			}
			cores = append(cores, fileCore)
		}
	}

	if len(cores) == 0 {
		// 默认控制台输出
		consoleEncoder := zapcore.NewConsoleEncoder(buildConsoleEncoderConfig(config))
		consoleWriter := zapcore.Lock(os.Stdout)
		return zapcore.NewCore(consoleEncoder, consoleWriter, ParseLevel(config.Level)), nil
	}

	return zapcore.NewTee(cores...), nil
}

// createFileCore 创建文件核心
func createFileCore(fileConfig FileConfig, rotationConfig RotationConfig, config *Config) (zapcore.Core, error) {
	// 创建编码器
	var encoder zapcore.Encoder
	if config.Format == "json" {
		encoder = zapcore.NewJSONEncoder(buildJSONEncoderConfig(config))
	} else {
		encoder = zapcore.NewConsoleEncoder(buildConsoleEncoderConfig(config))
	}

	// 创建写入器
	var writer zapcore.WriteSyncer
	if rotationConfig.Size.Enabled || rotationConfig.Time.Enabled {
		rotationWriter, err := NewRotationWriter(fileConfig, rotationConfig)
		if err != nil {
			return nil, err
		}
		writer = zapcore.AddSync(rotationWriter)
	} else {
		// 使用lumberjack进行基本的文件分割
		lumberjackLogger := &lumberjack.Logger{
			Filename:   fileConfig.Filename,
			MaxSize:    fileConfig.MaxSize,
			MaxBackups: fileConfig.MaxBackups,
			MaxAge:     fileConfig.MaxAge,
			Compress:   fileConfig.Compress,
			LocalTime:  fileConfig.LocalTime,
		}
		writer = zapcore.AddSync(lumberjackLogger)
	}

	// 设置日志级别过滤
	level := ParseLevel(config.Level)
	if fileConfig.LevelFilter != "" {
		level = ParseLevel(fileConfig.LevelFilter)
	}

	return zapcore.NewCore(encoder, writer, level), nil
}

// buildConsoleEncoderConfig 构建控制台编码器配置
func buildConsoleEncoderConfig(config *Config) zapcore.EncoderConfig {
	encoderConfig := zap.NewDevelopmentEncoderConfig()

	// 时间格式
	if config.Output.Console.TimeFormat != "" {
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(config.Output.Console.TimeFormat)
	} else {
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// 日志级别
	if config.Output.Console.Colorful {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	// 调用者信息
	if config.Caller.FullPath {
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	} else {
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return encoderConfig
}

// buildJSONEncoderConfig 构建JSON编码器配置
func buildJSONEncoderConfig(config *Config) zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()

	// 时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 日志级别
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// 调用者信息
	if config.Caller.FullPath {
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	} else {
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return encoderConfig
}

// CleanupOldLogs 清理旧日志文件
func CleanupOldLogs(logDir string, maxAge int) error {
	if maxAge <= 0 {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -maxAge)

	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查文件修改时间
		if info.ModTime().Before(cutoff) {
			// 删除过期文件
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove old log file %s: %w", path, err)
			}
		}

		return nil
	})
}

// CompressOldLogs 压缩旧日志文件
func CompressOldLogs(logDir string, delayHours int) error {
	if delayHours <= 0 {
		return nil
	}

	cutoff := time.Now().Add(-time.Duration(delayHours) * time.Hour)

	return filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录和已压缩文件
		if info.IsDir() || strings.HasSuffix(path, ".gz") {
			return nil
		}

		// 检查文件修改时间
		if info.ModTime().Before(cutoff) {
			// 压缩文件
			if err := compressFile(path); err != nil {
				return fmt.Errorf("failed to compress log file %s: %w", path, err)
			}
		}

		return nil
	})
}

// compressFile 压缩单个文件
func compressFile(filename string) error {
	// 这里可以实现gzip压缩逻辑
	// 为了简化，这里只是重命名文件
	compressedName := filename + ".gz"
	return os.Rename(filename, compressedName)
}

// GetLogFiles 获取日志文件列表
func GetLogFiles(logDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".log") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 按修改时间排序
	sort.Slice(files, func(i, j int) bool {
		info1, _ := os.Stat(files[i])
		info2, _ := os.Stat(files[j])
		return info1.ModTime().After(info2.ModTime())
	})

	return files, nil
}
