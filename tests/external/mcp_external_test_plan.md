# MCP 集成 - 外部测试计划

## 测试目标

通过外部测试（不依赖 gotest）验证 MCP 功能的完整性和正确性，包括：
- SSE 连接建立
- 认证机制
- 工具调用
- 错误处理
- 并发场景

## 测试环境要求

1. **已构建的二进制文件**：`target/tlcpchan`
2. **配置文件**：`config/config.yaml`（启用 MCP）
3. **测试工具**：curl, curl-sse-event, 或 Python/Node.js MCP 客户端
4. **网络端口**：
   - API 服务：`:20080`（默认）
   - MCP SSE 端点：`http://localhost:20080/api/mcp/sse`

## 测试配置准备

创建测试配置文件 `test-config-mcp.yaml`：

```yaml
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
```

## 测试套件

### 1. 基础功能测试 (BFT-001)

**目标**：验证 MCP 服务基本功能

| 测试 ID | 测试名称 | 测试方法 | 预期结果 |
|---------|---------|---------|---------|
| BFT-001-001 | 服务启动 | 启动 tlcpchan 服务 | 服务成功监听 :20080 |
| BFT-001-002 | SSE 端点可达 | curl -v http://localhost:20080/api/mcp/sse | HTTP 200，SSE 流响应 |
| BFT-001-003 | 未认证连接失败 | 不带 Authorization 头连接 | HTTP 401 |
| BFT-001-004 | 错误 API Key | Authorization: Bearer wrong-key | HTTP 401 |
| BFT-001-005 | 正确 API Key | Authorization: Bearer test-api-key-12345678 | HTTP 200，连接成功 |

**验证命令**：
```bash
# 启动服务
./target/tlcpchan -c test-config-mcp.yaml > /tmp/mcp-test.log 2>&1 &

# 等待服务启动
sleep 3

# 测试 SSE 端点（带认证）
curl -v -H "Authorization: Bearer test-api-key-12345678" \
  http://localhost:20080/api/mcp/sse 2>&1 | head -20

# 测试未认证连接（应该失败）
curl -v http://localhost:20080/api/mcp/sse 2>&1 | head -10
```

### 2. 工具调用测试 (TCT-002)

**目标**：验证所有 MCP 工具可以正确调用

#### 2.1 配置管理工具测试 (TCT-002-1)

| 工具名称 | 测试输入 | 预期输出 | 验证方法 |
|---------|---------|---------||---------|
| get_config | 无 | 当前配置 JSON | 验证 server.api.address 存在 |
| update_config | 新配置对象 | 更新后的配置 | 调用 get_config 验证 |
| reload_config | 无 | 重新加载后的配置 | 验证配置文件时间戳更新 |

**测试脚本示例**：
```bash
# 使用 curl 发送 MCP 工具调用（JSON-RPC 2.0）
cat > /tmp/test_get_config.json <<'EOF'
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_config",
    "arguments": {}
  }
}
EOF

curl -X POST \
  -H "Authorization: Bearer test-api-key-12345678" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_get_config.json \
  http://localhost:20080/api/mcp/sse
```

#### 2.2 实例管理工具测试 (TCT-002-2)

| 工具名称 | 测试输入 | 预期输出 | 验证方法 |
|---------|---------|---------|---------|
| list_instances | 无 | 实例列表 | 返回数组，每个实例有 name 和 status |
| get_instance | {"name": "test-inst"} | 实例详情 | 验证 name 匹配 |
| create_instance | 实例配置 | 创建的实例 | list_instances 验证存在 |
| delete_instance | {"name": "test-inst"} | 删除确认 | list_instances 验证不存在 |
| start_instance | {"name": "existing-inst"} | 启动状态 | 验证 status 为 "running" |
| stop_instance | {"name": "running-inst"} | 停止状态 | 验证 status 为 "stopped" |
| restart_instance | {"name": "inst"} | 重启状态 | 验证服务重启成功 |

**创建测试实例脚本**：
```bash
cat > /tmp/test_create_instance.json <<'EOF'
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "create_instance",
    "arguments": {
      "config": {
        "name": "mcp-test-instance",
        "type": "client",
        "protocol": "auto",
        "listen": ":9999",
        "target": "127.0.0.1:80",
        "enabled": false
      }
    }
  }
}
EOF

curl -X POST \
  -H "Authorization: Bearer test-api-key-12345678" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_create_instance.json \
  http://localhost:20080/api/mcp/sse
```

#### 2.3 密钥管理工具测试 (TCT-002-3)

| 工具名称 | 测试输入 | 预期输出 | 验证方法 |
|---------|---------|---------|---------|
| list_keystores | 无 | 密钥存储列表 | 返回数组 |
| get_keystore | {"name": "default"} | 密钥存储详情 | 验证 name 匹配 |
| create_keystore | 密钥存储配置 | 创建的密钥存储 | list_keystores 验证 |
| update_keystore | 更新配置 | 更新后的密钥存储 | get_keystore 验证 |
| delete_keystore | {"name": "test-ks"} | 删除确认 | list_keystores 验证不存在 |

#### 2.4 系统信息工具测试 (TCT-002-4)

| 工具名称 | 测试输入 | 预期输出 | 验证方法 |
|---------|---------|---------|---------|
| get_system_info | 无 | 系统信息 | 验证 version, goVersion 字段 |
| get_system_stats | 无 | 系统统计 | 验证 cpuUsage, memoryUsage 字段 |

#### 2.5 日志管理工具测试 (TCT-002-5)

| 工具名称 | 测试输入 | 预期输出 | 验证方法 |
|---------|---------|---------|---------|
| get_system_logs | {"lines": 10} | 最近 10 行日志 | 返回数组，长度 ≤ 10 |
| get_system_logs | {"lines": 10, "level": "error"} | ERROR 级别日志 | 验证所有行包含 ERROR |

### 3. 错误处理测试 (EHT-003)

**目标**：验证错误场景的正确处理

| 测试 ID | 错误场景 | 预期响应 | 验证方法 |
|---------|---------|---------|---------|
| EHT-003-001 | 无效的工具名称 | error.code = -32601 | JSON-RPC 方法未找到 |
| EHT-003-002 | 缺少必需参数 | error.code = -32602 | 无效参数错误 |
| EHT-003-003 | 实例不存在 | HTTP 错误或 JSON-RPC 错误 | 错误消息包含"不存在" |
| EHT-003-004 | 端口冲突 | 创建实例时返回错误 | 错误消息包含"端口" |
| EHT-003-005 | 超时测试 | 长时间操作超时 | 30秒内返回超时错误 |

**错误处理测试脚本**：
```bash
# 测试无效工具
cat > /tmp/test_invalid_tool.json <<'EOF'
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "non_existent_tool",
    "arguments": {}
  }
}
EOF

curl -X POST \
  -H "Authorization: Bearer test-api-key-12345678" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_invalid_tool.json \
  http://localhost:20080/api/mcp/sse 2>&1 | grep -A 5 "error"
```

### 4. 并发测试 (CT-004)

**目标**：验证 MCP 服务的并发处理能力

| 测试 ID | 测试场景 | 并发数 | 预期结果 | 验证方法 |
|---------|---------|--------|---------|---------|
| CT-004-001 | 多个工具调用 | 5 | 所有调用成功 | 检查响应 ID 都匹配 |
| CT-004-002 | 同一工具并发调用 | 3 | 串行或正确处理 | 系统一致性验证 |
| CT-004-003 | 多客户端连接 | 3 | 所有连接成功 | 监听连接数 |

**并发测试脚本示例**：
```bash
#!/bin/bash
# 并发调用多个工具

for i in {1..5}; do
  (
    cat > /tmp/concurrent_$i.json <<EOF
{
  "jsonrpc": "2.0",
  "id": $i,
  "method": "tools/call",
  "params": {
    "name": "get_system_info",
    "arguments": {}
  }
}
EOF

    curl -X POST \
      -H "Authorization: Bearer test-api-key-12345678" \
      -H "Content-Type: application/json" \
      -d @/tmp/concurrent_$i.json \
      http://localhost:20080/api/mcp/sse -s > /tmp/response_$i.json 2>&1
  ) &
done

wait

# 验证所有响应
success_count=0
for i in {1..5}; do
  if grep -q '"id":'$i /tmp/response_$i.json; then
    success_count=$((success_count + 1))
  fi
done

echo "并发测试结果: $success_count/5 成功"
```

### 5. 性能测试 (PT-005)

**目标**：验证 MCP 服务的性能指标

| 测试 ID | 测试场景 | 性能指标 | 目标值 | 验证方法 |
|---------|---------|---------|---------|---------|
| PT-005-001 | 连接建立时间 | Time to first byte | < 500ms | 使用 time 命令测量 |
| PT-005-002 | 工具调用响应 | 响应时间 | < 100ms | 测量 get_system_info |
| PT-005-003 | 并发吞吐量 | 请求/秒 | > 10 RPS | 压力测试 |
| PT-005-004 | 内存使用 | RSS 增长 | < 50MB | 监控进程内存 |

**性能测试脚本**：
```bash
#!/bin/bash
# 工具调用响应时间测试

for i in {1..10}; do
  start=$(date +%s%N)
  curl -X POST \
    -H "Authorization: Bearer test-api-key-12345678" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_system_info","arguments":{}}}' \
    http://localhost:20080/api/mcp/sse -s -o /dev/null
  end=$(date +%s%N)
  elapsed=$(( (end - start) / 1000000 ))
  echo "调用 $i: ${elapsed}ms"
done
```

### 6. 长连接稳定性测试 (LCT-006)

**目标**：验证 SSE 长连接的稳定性

| 测试 ID | 测试场景 | 持续时间 | 预期结果 | 验证方法 |
|---------|---------|---------|---------|---------|
| LCT-006-001 | 空闲连接保持 | 60 秒 | 连接不中断 | 监控连接状态 |
| LCT-006-002 | 周期性心跳 | 120 秒 | 正常接收事件 | 验证事件流 |
| LCT-006-003 | 服务重启后重连 | 服务重启 | 客户端重连成功 | 自动重连验证 |

**长连接测试脚本**：
```bash
#!/bin/bash
# SSE 长连接稳定性测试（使用 curl 事件流）

timeout 65s curl -v \
  -H "Authorization: Bearer test-api-key-12345678" \
  -H "Accept: text/event-stream" \
  http://localhost:20080/api/mcp/sse 2>&1 | tee /tmp/sse-stream.log

# 验证连接在 60 秒内保持活跃
if grep -q "data:" /tmp/sse-stream.log; then
  echo "✓ 长连接测试通过：成功接收 SSE 事件"
else
  echo "✗ 长连接测试失败：未接收到 SSE 事件"
fi
```

### 7. 安全性测试 (ST-007)

**目标**：验证 MCP 安全机制

| 测试 ID | 测试场景 | 预期结果 | 验证方法 |
|---------|---------|---------|---------|
| ST-007-001 | 无 API Key 连接 | HTTP 401 | 认证失败 |
| ST-007-002 | 错误的 API Key | HTTP 401 | 认证失败 |
| ST-007-003 | API Key 格式错误 | HTTP 401 | 认证失败 |
| ST-007-004 | 开放访问模式 | HTTP 200 | 配置为空 API Key 时允许 |

## 测试执行脚本

### 完整测试套件执行

```bash
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
    "$API_BASE/api/mcp/sse" | grep -q "200"; then
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
    "$API_BASE/api/mcp/sse")
  
  if echo "$response" | grep -q '"result"' && \
     echo "$response" | grep -q '"version"'; then
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
    "$API_BASE/api/mcp/sse")
  
  if echo "$response" | grep -q '"instances"'; then
    print_result "list_instances 工具" "PASS"
  else
    print_result "list_instances 工具" "FAIL"
    echo "响应: $response"
  fi
}

# 测试错误处理
test_error_handling() {
  echo -e "\n${YELLOW}=== 错误处理测试 ===${NC}"
  
  # 测试无效工具
  response=$(curl -s -X POST \
    -H "Authorization: Bearer $TEST_API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"invalid_tool","arguments":{}}}' \
    "$API_BASE/api/mcp/sse")
  
  if echo "$response" | grep -q '"error"'; then
    print_result "无效工具错误处理" "PASS"
  else
    print_result "无效工具错误处理" "FAIL"
    echo "响应: $response"
  fi
}

# 主函数
main() {
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
```

## 测试报告模板

```markdown
# MCP 外部测试报告

## 测试环境

- **操作系统**: Linux x86_64
- **TLCP Channel 版本**: 1.0.0
- **MCP SDK 版本**: v1.3.1
- **测试时间**: 2026-02-28
- **测试人员**: [填写]

## 测试结果概要

| 测试套件 | 总数 | 通过 | 失败 | 通过率 |
|---------|------|------|------|--------|
| 基础功能测试 | 5 | 5 | 0 | 100% |
| 工具调用测试 | 19 | 18 | 1 | 94.7% |
| 错误处理测试 | 5 | 5 | 0 | 100% |
| 并发测试 | 3 | 3 | 0 | 100% |
| 性能测试 | 4 | 4 | 0 | 100% |
| 长连接稳定性测试 | 3 | 2 | 1 | 66.7% |
| 安全性测试 | 4 | 4 | 0 | 100% |
| **总计** | **43** | **41** | **2** | **95.3%** |

## 失败测试详情

| 测试 ID | 测试名称 | 失败原因 | 严重程度 |
|---------|---------|---------|---------|
| TCT-002-015 | create_instance | 端口冲突检测不准确 | 中 |
| LCT-006-003 | 服务重启后重连 | 心跳丢失导致连接断开 | 低 |

## 性能指标

| 指标 | 测量值 | 目标值 | 状态 |
|------|--------|--------|------|
| 连接建立时间 | 120ms | < 500ms | ✓ |
| get_system_info 响应时间 | 45ms | < 100ms | ✓ |
| 并发吞吐量 | 25 RPS | > 10 RPS | ✓ |
| 内存使用增长 | 32MB | < 50MB | ✓ |

## 建议和改进

1. 修复端口冲突检测逻辑
2. 改进长连接心跳机制
3. 添加更多边界条件测试

## 结论

MCP 功能基本稳定，通过率 95.3%。建议修复失败问题后进行回归测试。
```

## 自动化测试集成

### CI/CD 集成

可以将外部测试集成到 GitHub Actions：

```yaml
name: MCP External Tests

on:
  push:
    paths:
      - 'tlcpchan/controller/mcp/**'
      - 'tlcpchan/config/**'
  pull_request:
    paths:
      - 'tlcpchan/controller/mcp/**'
      - 'tlcpchan/config/**'

jobs:
  external-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with
          go-version: '1.26'
      
      - name: Build
        run: ./build.sh
      
      - name: Run External Tests
        run: |
          chmod +x ./tests/external/run_mcp_tests.sh
          ./tests/external/run_mcp_tests.sh
      
      - name: Upload Test Results
        if: failure()
        uses: actions/upload-artifact@v3
        with:
          name: test-logs
          path: /tmp/tlcpchan-mcp-test.log
```

## 测试覆盖度分析

基于外部测试，以下功能已被覆盖：

### 已覆盖功能
- ✓ SSE 连接建立和管理
- ✓ API Key 认证
- ✓ 所有 19 个 MCP 工具的基本调用
- ✓ 错误响应格式
- ✓ 并发工具调用
- ✓ 基本性能指标

### 未覆盖功能
- ✗ 实例实际启动/停止验证
- ✗ 密钥库文件操作验证
- ✗ 证书添加/删除验证
- ✗ 极端并发压力测试
- ✗ 长时间运行稳定性（> 1小时）

## 后续测试计划

1. **Phase 2**: 添加真实的实例端到端测试
2. **Phase 3**: 集成实际密钥库和证书操作
3. **Phase 4**: 压力测试（100+ 并发连接）
4. **Phase 5**: 长期稳定性测试（24小时持续运行）
