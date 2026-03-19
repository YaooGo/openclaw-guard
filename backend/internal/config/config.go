package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	mu sync.RWMutex

	// 基础配置
	HomeDir      string
	ConfigDir    string
	DataDir      string
	LogDir       string

	// OpenClaw 相关路径
	OpenClawDir    string
	ClaudeCodeDir  string

	// 监控配置
	MonitorEnabled    bool
	WhitelistPaths    []string
	BlacklistPaths    []string
	DangerousCommands []string

	// 扫描配置
	ScanOnStart       bool
	AutoFix           bool
	ExportReportPath  string
}

var globalConfig *Config
var once sync.Once

func Load() *Config {
	once.Do(func() {
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".openclaw-guard")
		dataDir := filepath.Join(homeDir, ".openclaw-guard", "data")
		logDir := filepath.Join(homeDir, ".openclaw-guard", "logs")

		globalConfig = &Config{
			HomeDir:      homeDir,
			ConfigDir:    configDir,
			DataDir:      dataDir,
			LogDir:       logDir,
			OpenClawDir:  filepath.Join(homeDir, ".openclaw"),
			ClaudeCodeDir: filepath.Join(homeDir, ".claude"),

			MonitorEnabled: true,
			WhitelistPaths: []string{
				filepath.Join(homeDir, "Projects"),
				filepath.Join(homeDir, "code"),
				filepath.Join(homeDir, "Documents"),
				filepath.Join(homeDir, "Desktop"),
			},
			BlacklistPaths: []string{
				filepath.Join(homeDir, ".ssh"),
				filepath.Join(homeDir, ".gnupg"),
				"/etc",
				"/usr",
				"/System",
			},
			DangerousCommands: []string{
				"rm -rf",
				"rm -r /",
				"mkfs",
				"dd if=",
				":(){ :|:& };:",
				"chmod 000",
				"chown root",
				"format",
				"del /f",
			},

			ScanOnStart: true,
			AutoFix:     false,
		}

		// 创建必要目录
		os.MkdirAll(configDir, 0755)
		os.MkdirAll(dataDir, 0755)
		os.MkdirAll(logDir, 0755)
	})

	return globalConfig
}

func Get() *Config {
	return globalConfig
}

func (c *Config) AddWhitelistPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.WhitelistPaths = append(c.WhitelistPaths, path)
}

func (c *Config) RemoveWhitelistPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, p := range c.WhitelistPaths {
		if p == path {
			c.WhitelistPaths = append(c.WhitelistPaths[:i], c.WhitelistPaths[i+1:]...)
			break
		}
	}
}

func (c *Config) AddBlacklistPath(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BlacklistPaths = append(c.BlacklistPaths, path)
}

func (c *Config) SetMonitorEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.MonitorEnabled = enabled
}

func (c *Config) IsWhitelisted(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.WhitelistPaths {
		if isPathUnder(path, p) {
			return true
		}
	}
	return false
}

func (c *Config) IsBlacklisted(path string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, p := range c.BlacklistPaths {
		if isPathUnder(path, p) {
			return true
		}
	}
	return false
}

// isPathUnder 正确判断 path 是否在 parent 目录下（或本身就是 parent）
func isPathUnder(path, parent string) bool {
	// 规范化路径，去掉尾部分隔符
	parent = filepath.Clean(parent)
	path = filepath.Clean(path)
	if path == parent {
		return true
	}
	// 确保 parent 后紧跟分隔符，避免 /foo 匹配 /foobar
	return strings.HasPrefix(path, parent+string(filepath.Separator))
}

func (c *Config) GetConfig() (map[string]interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]interface{}{
		"whitelist_paths":    c.WhitelistPaths,
		"blacklist_paths":   c.BlacklistPaths,
		"dangerous_commands": c.DangerousCommands,
		"monitor_enabled":   c.MonitorEnabled,
		"scan_on_start":     c.ScanOnStart,
		"auto_fix":          c.AutoFix,
	}, nil
}

func (c *Config) SetWhitelistPaths(paths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.WhitelistPaths = paths
}

func (c *Config) SetBlacklistPaths(paths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BlacklistPaths = paths
}
