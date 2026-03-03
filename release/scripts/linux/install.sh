#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================"
echo "  TLCP Channel 安装脚本"
echo "========================================"

# 创建目录
echo "[INFO] 创建目录..."
mkdir -p /etc/tlcpchan
mkdir -p /etc/tlcpchan/keystores
mkdir -p /etc/tlcpchan/logs

# 复制文件
echo "[INFO] 复制文件..."
cp "$SCRIPT_DIR/tlcpchan" /etc/tlcpchan/
cp "$SCRIPT_DIR/tlcpchan-cli" /etc/tlcpchan/
if [ -d "$SCRIPT_DIR/ui" ]; then
    cp -r "$SCRIPT_DIR/ui" /etc/tlcpchan/
fi
if [ -d "$SCRIPT_DIR/rootcerts" ]; then
    cp -r "$SCRIPT_DIR/rootcerts" /etc/tlcpchan/
fi
if [ -f "$SCRIPT_DIR/tlcpchan.service" ]; then
    cp "$SCRIPT_DIR/tlcpchan.service" /usr/lib/systemd/system/
fi

# 复制默认配置文件（如果不存在）
echo "[INFO] 处理配置文件..."
if [ -f "$SCRIPT_DIR/config.yaml" ]; then
    if [ ! -f "/etc/tlcpchan/config.yaml" ]; then
        cp "$SCRIPT_DIR/config.yaml" /etc/tlcpchan/
        echo "[INFO] 已安装默认配置文件"
    else
        echo "[INFO] 检测到已有配置文件，跳过覆盖"
    fi
else
    echo "[WARN] 未找到默认配置文件"
fi

# 设置权限
echo "[INFO] 设置权限..."
chmod +x /etc/tlcpchan/tlcpchan
chmod +x /etc/tlcpchan/tlcpchan-cli

# 创建软链接
echo "[INFO] 创建软链接..."
ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 重新加载 systemd
echo "[INFO] 重新加载 systemd..."
systemctl daemon-reload 2>/dev/null || true

echo "========================================"
echo "  安装完成！"
echo "========================================"
echo "使用 'systemctl start tlcpchan' 启动服务"
echo "使用 'systemctl enable tlcpchan' 设置开机自启"
echo "使用 'tlcpchan -version' 查看版本"
