package ipc

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/openclaw-guard/backend/internal/config"
	"github.com/openclaw-guard/backend/internal/monitor"
	"github.com/openclaw-guard/backend/internal/scanner"
)

type Request struct {
	ID     string                 `json:"id"`
	Action string                 `json:"action"`
	Data   map[string]interface{} `json:"data"`
}

type Response struct {
	ID      string                 `json:"id"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error,omitempty"`
}

type Event struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

type Server struct {
	config  *config.Config
	scanner *scanner.Scanner
	monitor *monitor.Monitor
	mu      sync.RWMutex
}

func NewServer(cfg *config.Config, s *scanner.Scanner, m *monitor.Monitor) *Server {
	return &Server{
		config:  cfg,
		scanner: s,
		monitor: m,
	}
}

func (s *Server) Start() error {
	// 发送启动成功事件
	s.sendEvent("started", map[string]interface{}{
		"version": "1.0.0",
	})

	// 阻塞直到收到系统中断信号（SIGINT / SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	return nil
}

func (s *Server) HandleRequest(req Request) {
	var resp Response

	switch req.Action {
	case "scan":
		resp = s.handleScan(req)
	case "get_monitor_status":
		resp = s.handleGetMonitorStatus(req)
	case "set_monitor_enabled":
		resp = s.handleSetMonitorEnabled(req)
	case "get_config":
		resp = s.handleGetConfig(req)
	case "update_config":
		resp = s.handleUpdateConfig(req)
	case "get_logs":
		resp = s.handleGetLogs(req)
	case "export_report":
		resp = s.handleExportReport(req)
	default:
		resp = Response{
			ID:      req.ID,
			Success: false,
			Error:   "Unknown action: " + req.Action,
		}
	}

	s.sendResponse(resp)
}

func (s *Server) handleScan(req Request) Response {
	result, err := s.scanner.ScanAll()
	if err != nil {
		return Response{
			ID:      req.ID,
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"result": result,
		},
	}
}

func (s *Server) handleGetMonitorStatus(req Request) Response {
	status := s.monitor.GetStatus()

	return Response{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"status": status,
		},
	}
}

func (s *Server) handleSetMonitorEnabled(req Request) Response {
	enabled, ok := req.Data["enabled"].(bool)
	if !ok {
		return Response{
			ID:      req.ID,
			Success: false,
			Error:   "enabled must be a boolean",
		}
	}

	s.config.SetMonitorEnabled(enabled)
	if enabled {
		s.monitor.Start()
	} else {
		s.monitor.Stop()
	}

	return Response{
		ID:      req.ID,
		Success: true,
	}
}

func (s *Server) handleGetConfig(req Request) Response {
	cfg, err := s.config.GetConfig()
	if err != nil {
		return Response{
			ID:      req.ID,
			Success: false,
			Error:   err.Error(),
		}
	}

	return Response{
		ID:      req.ID,
		Success: true,
		Data:    cfg,
	}
}

func (s *Server) handleUpdateConfig(req Request) Response {
	// 更新白名单
	if whitelist, ok := req.Data["whitelist_paths"].([]interface{}); ok {
		paths := make([]string, 0, len(whitelist))
		for _, p := range whitelist {
			if str, ok := p.(string); ok {
				paths = append(paths, str)
			}
		}
		s.config.SetWhitelistPaths(paths)
	}

	// 更新黑名单
	if blacklist, ok := req.Data["blacklist_paths"].([]interface{}); ok {
		paths := make([]string, 0, len(blacklist))
		for _, p := range blacklist {
			if str, ok := p.(string); ok {
				paths = append(paths, str)
			}
		}
		s.config.SetBlacklistPaths(paths)
	}

	return Response{
		ID:      req.ID,
		Success: true,
	}
}

func (s *Server) handleGetLogs(req Request) Response {
	logs := s.monitor.GetLogs()

	return Response{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"logs": logs,
		},
	}
}

func (s *Server) handleExportReport(req Request) Response {
	result, err := s.scanner.ScanAll()
	if err != nil {
		return Response{
			ID:      req.ID,
			Success: false,
			Error:   err.Error(),
		}
	}

	// 生成报告
	report := s.scanner.GenerateReport(result)

	return Response{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"report": report,
		},
	}
}

func (s *Server) sendResponse(resp Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}

func (s *Server) sendEvent(eventType string, data map[string]interface{}) {
	event := Event{
		Event: eventType,
		Data:  data,
	}

	eventData, err := json.Marshal(event)
	if err != nil {
		return
	}
	fmt.Fprintln(os.Stdout, string(eventData))
}

func (s *Server) BroadcastAlert(alert monitor.Alert) {
	s.sendEvent("alert", map[string]interface{}{
		"type":      alert.Type,
		"severity":  alert.Severity,
		"message":   alert.Message,
		"operation": alert.Operation,
		"timestamp": alert.Timestamp,
	})
}
