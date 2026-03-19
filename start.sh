#!/bin/bash

# OpenClaw Guard 一键启动脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 OpenClaw Guard 启动中...${NC}"

# 检查并安装依赖
check_deps() {
    echo -e "${YELLOW}📦 检查依赖...${NC}"

    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go 未安装，请先安装 Go 1.23+${NC}"
        exit 1
    fi

    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js 未安装，请先安装 Node.js 18+${NC}"
        exit 1
    fi

    # 检查并安装前端依赖
    if [ ! -d "frontend/node_modules" ]; then
        echo -e "${YELLOW}  → 安装前端依赖...${NC}"
        cd frontend
        npm install
        cd ..
    fi

    # 检查并安装 Electron 依赖
    if [ ! -d "electron/node_modules" ]; then
        echo -e "${YELLOW}  → 安装 Electron 依赖...${NC}"
        cd electron
        npm install
        cd ..
    fi

    # 检查并安装后端依赖
    if [ ! -f "backend/go.sum" ]; then
        echo -e "${YELLOW}  → 安装后端依赖...${NC}"
        cd backend
        go mod download
        cd ..
    fi

    # 创建 bin 目录
    mkdir -p electron/bin

    echo -e "${GREEN}✅ 依赖检查完成${NC}"
}

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}🧹 正在清理...${NC}"
    if [ -n "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    echo -e "${GREEN}✅ 已退出${NC}"
    exit 0
}

# 捕获退出信号
trap cleanup SIGINT SIGTERM

# 清理占用端口的进程
cleanup_ports() {
    echo -e "${YELLOW}🔌 清理占用端口的进程...${NC}"
    # 杀掉占用 5173 端口的 node 进程
    lsof -ti:5173 | xargs kill -9 2>/dev/null || true
    lsof -ti:5174 | xargs kill -9 2>/dev/null || true
    lsof -ti:5175 | xargs kill -9 2>/dev/null || true
}

# 检查依赖
check_deps
cleanup_ports

# 启动前端开发服务器
echo -e "${YELLOW}🌐 启动前端开发服务器...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

# 等待前端服务启动
echo -e "${YELLOW}⏳ 等待前端服务就绪...${NC}"
sleep 3

# 启动 Electron
echo -e "${YELLOW}⚡ 启动 Electron...${NC}"
cd electron
NODE_ENV=development npm start
