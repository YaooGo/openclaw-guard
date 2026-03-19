package monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/openclaw-guard/backend/internal/config"
	"github.com/openclaw-guard/backend/pkg/rules"
)

type Monitor struct {
	config     *config.Config
	ruleEngine *rules.Engine
	running    bool
	mu         sync.RWMutex
	logs       []LogEntry
	alertChan  chan Alert
	stopChan   chan struct{} // 用于优雅停止 goroutine
	// 记录上一次持久化到的日志位置，避免重复写入
	persistedOffset int
}

type LogEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Operation string    `json:"operation"`
	Path      string    `json:"path,omitempty"`
	Command   string    `json:"command,omitempty"`
	Allowed   bool      `json:"allowed"`
	Reason    string    `json:"reason,omitempty"`
}

type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Operation string    `json:"operation"`
	Timestamp time.Time `json:"timestamp"`
}

type Operation struct {
	Type    string `json:"type"`    // read, write, delete, execute
	Path    string `json:"path"`
	Command string `json:"command,omitempty"`
	Source  string `json:"source"` // claude-code, cursor, etc.
}

type MonitorStatus struct {
	Running         bool      `json:"running"`
	TotalOperations int       `json:"total_operations"`
	BlockedCount    int       `json:"blocked_count"`
	AllowedCount    int       `json:"allowed_count"`
	LastActivity    time.Time `json:"last_activity"`
}

func New(cfg *config.Config, engine *rules.Engine) *Monitor {
	return &Monitor{
		config:     cfg,
		ruleEngine: engine,
		running:    false,
		logs:       make([]LogEntry, 0, 1000),
		alertChan:  make(chan Alert, 100),
		stopChan:   make(chan struct{}),
	}
}

func (m *Monitor) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return
	}

	m.running = true
	// 每次 Start 时重建 stopChan，保证干净的状态
	m.stopChan = make(chan struct{})

	// 启动日志记录服务
	go m.logService()

	// 启动告警处理
	go m.alertHandler()
}

func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}
	m.running = false
	// 关闭 stopChan，通知所有 goroutine 退出
	close(m.stopChan)
	// 关闭 alertChan，让 alertHandler goroutine 正常退出
	close(m.alertChan)
	// 重新初始化 alertChan 以备下次 Start 使用
	m.alertChan = make(chan Alert, 100)
}

func (m *Monitor) CheckOperation(op Operation) (bool, string) {
	m.mu.RLock()
	running := m.running
	m.mu.RUnlock()

	if !running {
		return true, "" // 监控未启用，放行
	}

	// 检查路径黑名单
	if op.Path != "" && m.config.IsBlacklisted(op.Path) {
		m.logOperation(op, false, "路径在黑名单中")
		m.sendAlert("danger", "high", fmt.Sprintf("阻止访问黑名单路径: %s", op.Path), op.Type)
		return false, "路径在黑名单中"
	}

	// 检查危险命令
	if op.Command != "" {
		if m.isDangerousCommand(op.Command) {
			m.logOperation(op, false, "检测到危险命令")
			m.sendAlert("danger", "high", fmt.Sprintf("检测到危险命令: %s", op.Command), op.Type)
			return false, "检测到危险命令"
		}
	}

	// 检查敏感操作
	if op.Type == "delete" {
		if !m.config.IsWhitelisted(op.Path) {
			m.logOperation(op, false, "删除操作需要白名单确认")
			m.sendAlert("warning", "medium", fmt.Sprintf("阻止删除非白名单文件: %s", op.Path), op.Type)
			return false, "删除操作需要白名单确认"
		}
	}

	// 检查写入操作
	if op.Type == "write" {
		// 对于写入操作，如果是白名单路径则放行，否则记录警告
		if !m.config.IsWhitelisted(op.Path) && op.Path != "" {
			m.logOperation(op, true, "写入非白名单路径")
			m.sendAlert("warning", "low", fmt.Sprintf("写入非白名单路径: %s", op.Path), op.Type)
			return true, ""
		}
	}

	// 默认放行读取操作
	m.logOperation(op, true, "")
	return true, ""
}

func (m *Monitor) isDangerousCommand(command string) bool {
	lowerCmd := strings.ToLower(command)
	for _, dangerous := range m.config.DangerousCommands {
		if strings.Contains(lowerCmd, strings.ToLower(dangerous)) {
			return true
		}
	}
	return false
}

func (m *Monitor) logOperation(op Operation, allowed bool, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := LogEntry{
		ID:        generateID(),
		Timestamp: time.Now(),
		Type:      op.Type,
		Operation: op.Command, // 命令操作记录命令，路径操作此字段为空
		Path:      op.Path,
		Allowed:   allowed,
		Reason:    reason,
	}

	m.logs = append(m.logs, entry)

	// 限制日志数量
	if len(m.logs) > 1000 {
		// 截掉头部 100 条时，同步调整已持久化偏移量
		trimmed := 100
		m.logs = m.logs[trimmed:]
		if m.persistedOffset > trimmed {
			m.persistedOffset -= trimmed
		} else {
			m.persistedOffset = 0
		}
	}
}

func (m *Monitor) GetStatus() MonitorStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blocked := 0
	allowed := 0
	lastActivity := time.Time{}

	for _, log := range m.logs {
		if log.Allowed {
			allowed++
		} else {
			blocked++
		}
		if log.Timestamp.After(lastActivity) {
			lastActivity = log.Timestamp
		}
	}

	return MonitorStatus{
		Running:         m.running,
		TotalOperations: len(m.logs),
		BlockedCount:    blocked,
		AllowedCount:    allowed,
		LastActivity:    lastActivity,
	}
}

func (m *Monitor) GetLogs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回最近100条
	if len(m.logs) <= 100 {
		result := make([]LogEntry, len(m.logs))
		copy(result, m.logs)
		return result
	}

	start := len(m.logs) - 100
	result := make([]LogEntry, 100)
	copy(result, m.logs[start:])
	return result
}

func (m *Monitor) logService() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.persistLogs()
		case <-m.stopChan:
			// 收到停止信号，退出 goroutine
			return
		}
	}
}

func (m *Monitor) persistLogs() {
	m.mu.Lock()
	// 只取尚未持久化的新日志
	newLogs := m.logs[m.persistedOffset:]
	if len(newLogs) == 0 {
		m.mu.Unlock()
		return
	}
	// 快照新日志，更新偏移量
	snapshot := make([]LogEntry, len(newLogs))
	copy(snapshot, newLogs)
	m.persistedOffset = len(m.logs)
	m.mu.Unlock()

	// 确保日志目录存在
	os.MkdirAll(m.config.LogDir, 0755)

	logFile := filepath.Join(m.config.LogDir, fmt.Sprintf("monitor-%s.json", time.Now().Format("2006-01-02")))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, log := range snapshot {
		encoder.Encode(log)
	}
}

func (m *Monitor) alertHandler() {
	for alert := range m.alertChan {
		// 这里可以通过 IPC 发送告警到前端
		// 暂时只记录到日志
		m.logAlert(alert)
	}
	// alertChan 关闭后，goroutine 正常退出
}

func (m *Monitor) logAlert(alert Alert) {
	alertLog := filepath.Join(m.config.LogDir, "alerts.log")
	file, err := os.OpenFile(alertLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	msg := fmt.Sprintf("[%s] %s: %s (操作: %s)\n",
		alert.Timestamp.Format("2006-01-02 15:04:05"),
		alert.Severity, alert.Message, alert.Operation)
	file.WriteString(msg)
}

func (m *Monitor) sendAlert(alertType, severity, message, operation string) {
	alert := Alert{
		ID:        generateID(),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Operation: operation,
		Timestamp: time.Now(),
	}

	select {
	case m.alertChan <- alert:
	default:
		// 告警队列满了，丢弃
	}
}

// generateID 使用 UUID 生成唯一 ID，避免高并发下 UnixNano 重复
func generateID() string {
	return uuid.New().String()
}
