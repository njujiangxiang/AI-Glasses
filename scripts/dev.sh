#!/usr/bin/env bash
# 一键启动前后端开发服务
# 用法: ./scripts/dev.sh          — 启动前后端
#       ./scripts/dev.sh backend  — 仅启动后端
#       ./scripts/dev.sh frontend — 仅启动前端

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PORT=8080
FRONTEND_PORT=5180

# 颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m'

PIDS=()

cleanup() {
  echo ""
  echo -e "${YELLOW}正在停止服务...${NC}"
  for pid in "${PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
  wait 2>/dev/null
  echo -e "${GREEN}所有服务已停止${NC}"
  exit 0
}

trap cleanup SIGINT SIGTERM

start_backend() {
  echo -e "${GREEN}[后端]${NC} 启动 Go API 服务 (port ${BACKEND_PORT})..."
  cd "$ROOT_DIR/server"
  go run ./cmd/api 2>&1 | sed "s/^/$(printf "${GREEN}[后端]${NC} ")/" &
  PIDS+=($!)
  cd "$ROOT_DIR"
}

start_frontend() {
  echo -e "${GREEN}[前端]${NC} 启动 Vite 开发服务 (port ${FRONTEND_PORT})..."
  cd "$ROOT_DIR/admin"
  if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}[前端]${NC} node_modules 不存在，正在安装依赖..."
    npm install
  fi
  node node_modules/vite/bin/vite.js --host 0.0.0.0 --port "$FRONTEND_PORT" 2>&1 | sed "s/^/$(printf "${GREEN}[前端]${NC} ")/" &
  PIDS+=($!)
  cd "$ROOT_DIR"
}

wait_for_service() {
  local name="$1"
  local port="$2"
  local max_wait=30
  local waited=0
  while ! curl -s -o /dev/null "http://127.0.0.1:${port}/" 2>/dev/null; do
    sleep 1
    waited=$((waited + 1))
    if [ "$waited" -ge "$max_wait" ]; then
      echo -e "${RED}[${name}]${NC} 启动超时 (${max_wait}s)"
      return 1
    fi
  done
  echo -e "${GREEN}[${name}]${NC} 已就绪 http://127.0.0.1:${port}"
}

# 解析参数
MODE="${1:-all}"

echo "========================================"
echo "  智能眼镜巡检任务管理系统 - 开发环境"
echo "========================================"
echo ""

case "$MODE" in
  backend)
    start_backend
    wait_for_service "后端" "$BACKEND_PORT"
    ;;
  frontend)
    start_frontend
    wait_for_service "前端" "$FRONTEND_PORT"
    ;;
  all|*)
    start_backend
    start_frontend
    wait_for_service "后端" "$BACKEND_PORT"
    wait_for_service "前端" "$FRONTEND_PORT"
    echo ""
    echo -e "${GREEN}✓ 全部服务已启动${NC}"
    echo -e "  后端 API:  http://127.0.0.1:${BACKEND_PORT}"
    echo -e "  前端页面:  http://localhost:${FRONTEND_PORT}"
    echo -e "  按 ${YELLOW}Ctrl+C${NC} 停止所有服务"
    ;;
esac

echo ""
wait
