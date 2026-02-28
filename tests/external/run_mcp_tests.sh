#!/bin/bash
# run_mcp_tests.sh - MCP 外部测试套件执行脚本

set -e

TLCPCHAN_BIN="./target/tlcpchan"
TEST_CONFIG="./config/test-config-mcp.yaml"
TEST_LOG="/tmp/tlcpchan-mcp-test.log"
TEST_API_KEY="test-api-key-12345678"
API_BASE="http://localhost:20080"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试结果统计
PASSED=0
FAILED=0

# 辅助函数
print_result() {
  local test_name="$1"
  local result="$2"
  if [ "$result" = "PASS" ]; then
    echo -e "${GREEN}✓ PASS${NC}: $test_name"
    ((PASSED++))
  else
    echo -e "${RED}✗ FAIL${NC}: $test_name"
    ((FAILED++))
  fi
}

# 启动服务
start_service() {
  echo "🚀 启动 TLCP Channel 服务..."
  $TLCPCHAN_BIN -c "$TEST_CONFIG" > "$TEST_LOG" 2>&1 &
  SERVICE_PID=$!
  # 等待服务启动
  for i in {1..10}; do
    if curl -s "$API_BASE/api/health" > /dev/null 2>&1; then
      echo "✓ 服务启动成功 (PID: $SERVICE_PID)"
      return 0
    fi
    sleep 1
  done

  echo "✗ 服务启动失败"
  return 1
}

# 停止服务
stop_service() {
  if [ -n "$SERVICE_PID" ]; then
    echo "🛑 停止服务 (PID: $SERVICE_PID)..."
    kill $SERVICE_PID 2>/dev/null || true
    wait $SERVICE_PID 2>/dev/null || true
  fi
}

# 测试基础功能
test_basic_functionality() {
  echo -e "\n${YELLOW}=== 基础功能测试 ===${NC}"
  
  # 测试 API Key 认证
  if curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TEST_API_KEY" \
    "$API_BASE/api/mcp/sse" | grep -q "200\|405"; then
    print_result "API Key 认证" "PASS"
  else
    print_result "API Key 认证" "FAIL"
  fi
  
  # 测试未认证访问
  if curl -s -o /dev/null -w "%{http_code}" \
    "$API_BASE/api/mcp/sse" | grep -q "401"; then
    print_result "未认证访问拒绝" "PASS"
  else
    print_result "未认证访问拒绝" "FAIL"
  fi
}

# 测试工具调用
test_tools() {
  echo -e "\n${YELLOW}=== 工具调用测试 ===${NC}"
  
  # 测试 get_system_info
  response=$(curl -s -X POST \
    -H "Authorization: Bearer $TEST_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_system_info","arguments":{}}}' \
    "$API_BASE/api/mcp/sse" 2>&1 | head -50)
  
  if echo "$response" | grep -q '"version"' && echo "$response" | grep -q '"goVersion"'; then
    print_result "get_system_info 工具" "PASS"
  else
    print_result "get_system_info 工具" "FAIL"
    echo "响应: $response"
  fi
  
  # 测试 list_instances
  response=$(curl -s -X POST \
    -H "Authorization: Bearer $TEST_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_instances","arguments":{}}}' \
    "$API_BASE/api/mcp/sse" 2>&1 | head -50)
  
  if echo "$response" | grep -q '"instances"'; then
    print_result "list_instances 工具" "PASS"
  else
    print_result "list_instances 工具" "FAIL"
    echo "响应: $response"
  fi

  # 测试 list_keystores
  response=$(curl -s -X POST \
    -H "Authorization: Bearer $TEST_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_keystores","arguments":{}}}' \
    "$API_BASE/api/mcp/sse" 2>&1 | head -50)
  
  if echo "$response" | grep -q '"keystores"'; then
    print_result "list_keystores 工具" "PASS"
  else
    print_result "list_keystores 工具" "FAIL"
  fi
}

# 测试错误处理
test_error_handling() {
  echo -e "\n${YELLOW}=== 错误处理测试 ===${NC}"
  
  # 测试无效工具
  response=$(curl -s -X POST \
    -H "Authorization: Bearer $TEST_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"non_existent_tool","arguments":{}}}' \
    "$API_BASE/api/mcp/sse" 2>&1 | head -50)
  
  if echo "$response" | grep -q '"error"' || echo "$response" | grep -q 'not found'; then
    print_result "无效工具错误处理" "PASS"
  else
    print_result "无效工具错误处理" "FAIL"
    echo "响应: $response"
  fi
}

# 主函数
main() {
  # 检查二进制文件
  if [ ! -f "$TLCPCHAN_BIN" ]; then
    echo "错误: 找不到 $TLCPCHAN_BIN"
    echo "请先运行 ./build.sh 构建项目"
    exit 1
  fi
  
  # 创建测试配置
  echo "📝 创建测试配置..."
  mkdir -p config
  cat > "$TEST_CONFIG" <<'EOF'
server:
  api:
    address: :20080
  log:
    level: debug
    file: /tmp/tlcpchan-mcp-test.log
    enabled: true

mcp:
  enabled: true
  api_key: "test-api-key-12345678"
  server_info:
    name: "tlcpchan-mcp-test"
    version: "1.0.0-test"
EOF
  
  # 启动服务
  if ! start_service; then
    echo "错误日志:"
    cat "$TEST_LOG" | tail -50
    exit 1
  fi
  
  # 设置清理
  trap stop_service EXIT
  
  # 运行测试
  test_basic_functionality
  test_tools
  test_error_handling
  
  # 输出结果
  echo -e "\n${YELLOW}=== 测试结果汇总 ===${NC}"
  echo -e "${GREEN}通过: $PASSED${NC}"
  echo -e "${RED}失败: $FAILED${NC}"
  if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过${NC}"
    exit 0
  else
    echo -e "${RED}✗ 部分测试失败${NC}"
    exit 1
  fi
}

# 运行主函数
main "$@"
