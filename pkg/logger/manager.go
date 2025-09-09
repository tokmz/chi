package logger

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Manager 日志管理器（优化版本：细粒度锁和性能改进）
type Manager struct {
	config ManagementConfig
	logDir string
	stopCh chan struct{}
	wg     sync.WaitGroup
	// 分离锁以减少锁竞争
	stateMu   sync.RWMutex // 状态锁
	fileMu    sync.RWMutex // 文件操作锁
	isRunning bool
	// 性能优化：缓存和批处理
	fileCache map[string]LogFileInfo
	cacheTime time.Time
	cacheTTL  time.Duration
	batchSize int
}

// LogFileInfo 日志文件信息
type LogFileInfo struct {
	Path      string
	Size      int64
	ModTime   time.Time
	IsGzipped bool
}

// NewManager 创建新的日志管理器
func NewManager(config ManagementConfig, logDir string) *Manager {
	return &Manager{
		config:    config,
		logDir:    logDir,
		stopCh:    make(chan struct{}),
		fileCache: make(map[string]LogFileInfo),
		cacheTTL:  5 * time.Minute, // 5分钟缓存TTL
		batchSize: 100,             // 批处理大小
	}
}

// Start 启动日志管理器（优化版本：减少锁持有时间）
func (m *Manager) Start() error {
	m.stateMu.Lock()
	if m.isRunning {
		m.stateMu.Unlock()
		return fmt.Errorf("manager is already running")
	}
	m.isRunning = true
	m.stateMu.Unlock()

	// 启动清理任务
	if m.config.Cleanup.Enabled {
		m.wg.Add(1)
		go m.cleanupWorker()
	}

	// 启动压缩任务
	if m.config.Compression.Enabled {
		m.wg.Add(1)
		go m.compressionWorker()
	}

	return nil
}

// Stop 停止日志管理器（优化版本：减少锁持有时间）
func (m *Manager) Stop() error {
	m.stateMu.Lock()
	if !m.isRunning {
		m.stateMu.Unlock()
		return nil
	}
	m.isRunning = false
	stopCh := m.stopCh
	m.stateMu.Unlock()

	close(stopCh)
	m.wg.Wait()

	// 重新创建 stopCh 以便下次启动
	m.stateMu.Lock()
	m.stopCh = make(chan struct{})
	m.stateMu.Unlock()

	return nil
}

// cleanupWorker 清理工作协程
func (m *Manager) cleanupWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.Cleanup.Interval)
	defer ticker.Stop()

	// 立即执行一次清理
	m.performCleanup()

	for {
		select {
		case <-ticker.C:
			m.performCleanup()
		case <-m.stopCh:
			return
		}
	}
}

// compressionWorker 压缩工作协程
func (m *Manager) compressionWorker() {
	defer m.wg.Done()

	ticker := time.NewTicker(time.Hour) // 每小时检查一次
	defer ticker.Stop()

	// 立即执行一次压缩
	m.performCompression()

	for {
		select {
		case <-ticker.C:
			m.performCompression()
		case <-m.stopCh:
			return
		}
	}
}

// performCleanup 执行清理操作
func (m *Manager) performCleanup() {
	if m.config.Cleanup.MaxAge <= 0 {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -m.config.Cleanup.MaxAge)

	files, err := m.getLogFiles()
	if err != nil {
		fmt.Printf("Failed to get log files: %v\n", err)
		return
	}

	for _, file := range files {
		if file.ModTime.Before(cutoff) {
			if err := os.Remove(file.Path); err != nil {
				fmt.Printf("Failed to remove old log file %s: %v\n", file.Path, err)
			} else {
				fmt.Printf("Removed old log file: %s\n", file.Path)
			}
		}
	}
}

// performCompression 执行压缩操作
func (m *Manager) performCompression() {
	if m.config.Compression.Delay <= 0 {
		return
	}

	cutoff := time.Now().Add(-time.Duration(m.config.Compression.Delay) * time.Hour)

	files, err := m.getLogFiles()
	if err != nil {
		fmt.Printf("Failed to get log files: %v\n", err)
		return
	}

	for _, file := range files {
		// 跳过已压缩的文件
		if file.IsGzipped {
			continue
		}

		// 跳过当前正在使用的日志文件
		if m.isCurrentLogFile(file.Path) {
			continue
		}

		if file.ModTime.Before(cutoff) {
			if err := m.compressFile(file.Path); err != nil {
				fmt.Printf("Failed to compress log file %s: %v\n", file.Path, err)
			} else {
				fmt.Printf("Compressed log file: %s\n", file.Path)
			}
		}
	}
}

// getLogFiles 获取日志文件列表（优化版本：使用缓存减少文件系统访问）
func (m *Manager) getLogFiles() ([]LogFileInfo, error) {
	m.fileMu.RLock()
	// 检查缓存是否有效
	if time.Since(m.cacheTime) < m.cacheTTL && len(m.fileCache) > 0 {
		files := make([]LogFileInfo, 0, len(m.fileCache))
		for _, file := range m.fileCache {
			files = append(files, file)
		}
		m.fileMu.RUnlock()

		// 按修改时间排序
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime.After(files[j].ModTime)
		})
		return files, nil
	}
	m.fileMu.RUnlock()

	// 缓存过期，重新扫描文件
	var files []LogFileInfo
	newCache := make(map[string]LogFileInfo)

	err := filepath.Walk(m.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 检查是否为日志文件
		if m.isLogFile(path) {
			fileInfo := LogFileInfo{
				Path:      path,
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				IsGzipped: strings.HasSuffix(path, ".gz"),
			}
			files = append(files, fileInfo)
			newCache[path] = fileInfo
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 更新缓存
	m.fileMu.Lock()
	m.fileCache = newCache
	m.cacheTime = time.Now()
	m.fileMu.Unlock()

	// 按修改时间排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	return files, nil
}

// isLogFile 检查是否为日志文件
func (m *Manager) isLogFile(path string) bool {
	ext := filepath.Ext(path)
	if ext == ".log" {
		return true
	}
	if ext == ".gz" && strings.HasSuffix(strings.TrimSuffix(path, ".gz"), ".log") {
		return true
	}
	return false
}

// isCurrentLogFile 检查是否为当前正在使用的日志文件
func (m *Manager) isCurrentLogFile(path string) bool {
	// 简单检查：如果文件在最近1分钟内被修改，认为是当前文件
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) < time.Minute
}

// compressFile 压缩文件
func (m *Manager) compressFile(filename string) error {
	switch m.config.Compression.Algorithm {
	case "gzip":
		return m.compressWithGzip(filename)
	case "lz4":
		return m.compressWithLZ4(filename)
	default:
		return m.compressWithGzip(filename)
	}
}

// compressWithGzip 使用gzip压缩文件
func (m *Manager) compressWithGzip(filename string) error {
	// 打开原文件
	srcFile, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 创建压缩文件
	gzFilename := filename + ".gz"
	gzFile, err := os.Create(gzFilename)
	if err != nil {
		return fmt.Errorf("failed to create gzip file: %w", err)
	}
	defer gzFile.Close()

	// 创建gzip写入器
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	// 复制数据
	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	// 关闭gzip写入器
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}

	// 删除原文件
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to remove original file: %w", err)
	}

	return nil
}

// compressWithLZ4 使用LZ4压缩文件（简化实现）
func (m *Manager) compressWithLZ4(filename string) error {
	// 这里可以实现LZ4压缩逻辑
	// 为了简化，这里只是重命名文件
	lz4Filename := filename + ".lz4"
	return os.Rename(filename, lz4Filename)
}

// GetStats 获取日志统计信息
func (m *Manager) GetStats() (*LogStats, error) {
	files, err := m.getLogFiles()
	if err != nil {
		return nil, err
	}

	stats := &LogStats{
		TotalFiles: len(files),
	}

	for _, file := range files {
		stats.TotalSize += file.Size
		if file.IsGzipped {
			stats.CompressedFiles++
			stats.CompressedSize += file.Size
		} else {
			stats.UncompressedFiles++
			stats.UncompressedSize += file.Size
		}
	}

	return stats, nil
}

// LogStats 日志统计信息
type LogStats struct {
	TotalFiles        int   `json:"total_files"`
	TotalSize         int64 `json:"total_size"`
	CompressedFiles   int   `json:"compressed_files"`
	CompressedSize    int64 `json:"compressed_size"`
	UncompressedFiles int   `json:"uncompressed_files"`
	UncompressedSize  int64 `json:"uncompressed_size"`
}

// ForceCleanup 强制执行清理
func (m *Manager) ForceCleanup() error {
	m.performCleanup()
	return nil
}

// ForceCompression 强制执行压缩
func (m *Manager) ForceCompression() error {
	m.performCompression()
	return nil
}

// CleanupByPattern 按模式清理文件
func (m *Manager) CleanupByPattern(pattern string, maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)

	matches, err := filepath.Glob(filepath.Join(m.logDir, pattern))
	if err != nil {
		return fmt.Errorf("failed to match pattern: %w", err)
	}

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			if err := os.Remove(match); err != nil {
				fmt.Printf("Failed to remove file %s: %v\n", match, err)
			} else {
				fmt.Printf("Removed file: %s\n", match)
			}
		}
	}

	return nil
}

// ArchiveLogs 归档日志文件
func (m *Manager) ArchiveLogs(archiveDir string, maxAge time.Duration) error {
	// 确保归档目录存在
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	files, err := m.getLogFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.ModTime.Before(cutoff) {
			// 移动文件到归档目录
			archivePath := filepath.Join(archiveDir, filepath.Base(file.Path))
			if err := os.Rename(file.Path, archivePath); err != nil {
				fmt.Printf("Failed to archive file %s: %v\n", file.Path, err)
			} else {
				fmt.Printf("Archived file: %s -> %s\n", file.Path, archivePath)
			}
		}
	}

	return nil
}

// IsRunning 检查管理器是否正在运行
func (m *Manager) IsRunning() bool {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()
	return m.isRunning
}
