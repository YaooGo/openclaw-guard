package scanner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/openclaw-guard/backend/internal/config"
	"github.com/openclaw-guard/backend/pkg/rules"
)

type Scanner struct {
	config     *config.Config
	ruleEngine *rules.Engine
}

type ScanResult struct {
	Timestamp   time.Time              `json:"timestamp"`
	Risks       []Risk                 `json:"risks"`
	Suggestions []Suggestion           `json:"suggestions"`
	Summary     ScanSummary            `json:"summary"`
	Details     map[string]interface{} `json:"details"`
}

type Risk struct {
	ID          string    `json:"id"`
	Level       string    `json:"level"`       // high, medium, low
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Path        string    `json:"path,omitempty"`
	Fixed       bool      `json:"fixed"`
	FixAction   string    `json:"fix_action,omitempty"`
}

type Suggestion struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type ScanSummary struct {
	TotalRisks       int `json:"total_risks"`
	HighRisks        int `json:"high_risks"`
	MediumRisks      int `json:"medium_risks"`
	LowRisks         int `json:"low_risks"`
	TotalSuggestions int `json:"total_suggestions"`
}

func New(cfg *config.Config, engine *rules.Engine) *Scanner {
	return &Scanner{
		config:     cfg,
		ruleEngine: engine,
	}
}

func (s *Scanner) ScanAll() (*ScanResult, error) {
	result := &ScanResult{
		Timestamp: time.Now(),
		Risks:     make([]Risk, 0),
		Suggestions: make([]Suggestion, 0),
		Details:   make(map[string]interface{}),
	}

	// 1. 扫描配置文件
	s.scanConfigFiles(result)

	// 2. 扫描敏感信息
	s.scanSensitiveInfo(result)

	// 3. 扫描文件权限
	s.scanFilePermissions(result)

	// 4. 扫描历史日志
	s.scanHistoryLogs(result)

	// 5. 生成统计摘要
	s.generateSummary(result)

	return result, nil
}

func (s *Scanner) scanConfigFiles(result *ScanResult) {
	configPaths := []string{
		filepath.Join(s.config.HomeDir, ".claude", "config.json"),
		filepath.Join(s.config.HomeDir, ".openclaw", "config.json"),
		filepath.Join(s.config.HomeDir, ".claude", "settings.json"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			s.checkConfigFile(path, result)
		}
	}
}

func (s *Scanner) checkConfigFile(path string, result *ScanResult) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		return
	}
	content := string(raw)

	// 按行检查明文 API Key
	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		if matched, _ := regexp.MatchString(`(?i)(api[_-]?key|token|secret)\s*[:=]\s*["\']?([a-zA-Z0-9]{20,})`, line); matched {
			result.Risks = append(result.Risks, Risk{
				ID:          fmt.Sprintf("config_apikey_%d", lineNum+1),
				Level:       "high",
				Title:       "\u914d\u7f6e\u6587\u4ef6\u5305\u542b\u660e\u6587 API Key",
				Description: fmt.Sprintf("\u5728 %s \u7b2c %d \u884c\u53d1\u73b0\u53ef\u80fd\u7684 API Key", path, lineNum+1),
				Path:        path,
				FixAction:   "\u5efa\u8bae\u4f7f\u7528\u73af\u5883\u53d8\u91cf\u6216\u5bc6\u9490\u7ba1\u7406\u5de5\u5177",
			})
		}
	}

	// 检查是否有日志记录敏感信息
	if strings.Contains(content, "logPath") || strings.Contains(content, "historyPath") {
		result.Suggestions = append(result.Suggestions, Suggestion{
			ID:          "config_log_cleanup",
			Title:       "\u542f\u7528\u81ea\u52a8\u65e5\u5fd7\u6e05\u7406",
			Description: "\u914d\u7f6e\u65e5\u5fd7\u81ea\u52a8\u6e05\u7406\u53ef\u4ee5\u51cf\u5c11\u654f\u611f\u4fe1\u606f\u6cc4\u9732\u98ce\u9669",
			Priority:    "medium",
		})
	}
}

func (s *Scanner) scanSensitiveInfo(result *ScanResult) {
	// 扫描常见敏感文件
	sensitiveFiles := []string{
		filepath.Join(s.config.HomeDir, ".ssh", "id_rsa"),
		filepath.Join(s.config.HomeDir, ".ssh", "id_ed25519"),
		filepath.Join(s.config.HomeDir, ".gnupg"),
	}

	for _, file := range sensitiveFiles {
		if info, err := os.Stat(file); err == nil {
			// 检查文件权限
			if info.Mode().Perm()&0077 != 0 {
				result.Risks = append(result.Risks, Risk{
					ID:          fmt.Sprintf("perm_%s", filepath.Base(file)),
					Level:       "high",
					Title:       "敏感文件权限过于宽松",
					Description: fmt.Sprintf("%s 可被其他用户读取", file),
					Path:        file,
					FixAction:   "运行: chmod 600 " + file,
				})
			}
		}
	}
}

func (s *Scanner) scanFilePermissions(result *ScanResult) {
	// 检查 .claude 目录权限
	claudeDir := filepath.Join(s.config.HomeDir, ".claude")
	if info, err := os.Stat(claudeDir); err == nil {
		if info.Mode().Perm()&0077 != 0 {
			result.Risks = append(result.Risks, Risk{
				ID:          "perm_claude_dir",
				Level:       "medium",
				Title:       "Claude 配置目录权限过于宽松",
				Description: ".claude 目录可被其他用户访问",
				Path:        claudeDir,
				FixAction:   "运行: chmod 700 ~/.claude",
			})
		}
	}
}

func (s *Scanner) scanHistoryLogs(result *ScanResult) {
	logPaths := []string{
		filepath.Join(s.config.HomeDir, ".claude", "history.json"),
		filepath.Join(s.config.HomeDir, ".openclaw", "logs"),
	}

	for _, path := range logPaths {
		s.scanLogFile(path, result)
	}
}

func (s *Scanner) scanLogFile(path string, result *ScanResult) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		return
	}

	lines := strings.Split(string(raw), "\n")
	sensitivePaths := []string{".ssh", ".gnupg", "AWS_ACCESS", "SECRET_KEY"}

	for lineNum, line := range lines {
		for _, sp := range sensitivePaths {
			if strings.Contains(line, sp) {
				result.Risks = append(result.Risks, Risk{
					ID:          fmt.Sprintf("log_sensitive_%d", lineNum+1),
					Level:       "medium",
					Title:       "日志包含敏感信息",
					Description: fmt.Sprintf("历史日志第 %d 行包含敏感路径或密钥", lineNum+1),
					Path:        path,
					FixAction:   "建议清理历史日志或启用日志脱敏",
				})
				break
			}
		}
	}
}


func (s *Scanner) generateSummary(result *ScanResult) {
	result.Summary.TotalRisks = len(result.Risks)
	result.Summary.TotalSuggestions = len(result.Suggestions)

	for _, risk := range result.Risks {
		switch risk.Level {
		case "high":
			result.Summary.HighRisks++
		case "medium":
			result.Summary.MediumRisks++
		case "low":
			result.Summary.LowRisks++
		}
	}
}

func (s *Scanner) GenerateReport(result *ScanResult) string {
	var sb strings.Builder

	sb.WriteString("OpenClaw 安全扫描报告\n")
	sb.WriteString("========================\n\n")
	sb.WriteString(fmt.Sprintf("扫描时间: %s\n\n", result.Timestamp.Format("2006-01-02 15:04:05")))

	sb.WriteString("风险摘要\n")
	sb.WriteString("--------\n")
	sb.WriteString(fmt.Sprintf("总计: %d 个风险\n", result.Summary.TotalRisks))
	sb.WriteString(fmt.Sprintf("  🔴 高危: %d\n", result.Summary.HighRisks))
	sb.WriteString(fmt.Sprintf("  🟡 中危: %d\n", result.Summary.MediumRisks))
	sb.WriteString(fmt.Sprintf("  🟢 低危: %d\n\n", result.Summary.LowRisks))

	if len(result.Risks) > 0 {
		sb.WriteString("详细风险\n")
		sb.WriteString("--------\n")
		for _, risk := range result.Risks {
			switch risk.Level {
			case "high":
				sb.WriteString("🔴 ")
			case "medium":
				sb.WriteString("🟡 ")
			case "low":
				sb.WriteString("🟢 ")
			}
			sb.WriteString(risk.Title + "\n")
			sb.WriteString("   " + risk.Description + "\n")
			if risk.FixAction != "" {
				sb.WriteString("   修复: " + risk.FixAction + "\n")
			}
			sb.WriteString("\n")
		}
	}

	if len(result.Suggestions) > 0 {
		sb.WriteString("优化建议\n")
		sb.WriteString("--------\n")
		for i, sug := range result.Suggestions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, sug.Title))
			sb.WriteString("   " + sug.Description + "\n\n")
		}
	}

	return sb.String()
}
