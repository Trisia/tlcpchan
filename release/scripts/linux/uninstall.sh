#!/bin/bash
set -e

echo "========================================"
echo "  TLCP Channel 卸载脚本"
echo "========================================"

# 停止服务
echo "[INFO] 停止服务..."
if systemctl is-active --quiet tlcpchan 2>/dev/null; then
    systemctl stop tlcpchan
fi

# 禁用服务
echo "[INFO] 禁用服务..."
if systemctl is-enabled --quiet tlcpchan 2>/dev/null; then
    systemctl disable tlcpchan
fi

# 重新加载 systemd
echo "[INFO] 重新加载 systemd..."
systemctl daemon-reload 2>/dev/null || true

# 删除软链接
echo "[INFO] 删除软链接..."
rm -f /usr/bin/tlcpchan
rm -f /usr/bin/tlcpchan-cli
rm -f /usr/bin/tlcpchan-ui
rm -f /usr/bin/tlcpc

# 删除 systemd 服务文件
echo "[INFO] 删除 systemd 服务文件..."
rm -f /usr/lib/systemd/system/tlcpchan.service

# 询问是否删除数据
read -p "是否删除配置和数据目录 /etc/tlcpchan? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[INFO] 删除 /etc/tlcpchan..."
    rm -rf /etc/tlcpchan
fi

# 询问是否删除用户
read -p "是否删除 tlcpchan 用户? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "[INFO] 删除 tlcpchan 用户..."
    userdel tlcpchan 2>/dev/null || true
fi

echo "========================================"
echo "  卸载完成！"
echo "========================================"
