package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/openclaw-guard/backend/internal/config"
	"github.com/openclaw-guard/backend/internal/ipc"
	"github.com/openclaw-guard/backend/internal/monitor"
	"github.com/openclaw-guard/backend/internal/scanner"
	"github.com/openclaw-guard/backend/pkg/rules"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化规则引擎
	ruleEngine := rules.NewEngine(cfg)

	// 初始化扫描器
	scanner := scanner.New(cfg, ruleEngine)

	// 初始化监控器
	monitor := monitor.New(cfg, ruleEngine)

	// 启动 IPC 服务
	server := ipc.NewServer(cfg, scanner, monitor)

	// 启动监控
	monitor.Start()
	defer monitor.Stop()

	// 从标准输入读取请求
	go handleInput(server, os.Stdin)

	// 启动服务
	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Service error: %v\n", err)
		os.Exit(1)
	}
}

func handleInput(server *ipc.Server, input io.Reader) {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		var req ipc.Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}
		server.HandleRequest(req)
	}
}


