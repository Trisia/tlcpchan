#!/bin/bash
set -e

# 创建 tlcpchan 用户
if ! getent passwd tlcpchan > /dev/null; then
    useradd -r -s /bin/false -d /etc/tlcpchan tlcpchan
fi

# 设置权限
chown -R tlcpchan:tlcpchan /etc/tlcpchan/keystores 2>/dev/null || true
chown -R tlcpchan:tlcpchan /etc/tlcpchan/logs 2>/dev/null || true

# 创建软链接到 /usr/bin
ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 处理默认配置文件（仅在全新安装时创建）
if [ "$1" = "1" ]; then
    if [ ! -f "/etc/tlcpchan/config.yaml" ] && [ -f "/etc/tlcpchan/config.yaml.rpmnew" ]; then
        mv "/etc/tlcpchan/config.yaml.rpmnew" "/etc/tlcpchan/config.yaml"
        echo "[INFO] 已安装默认配置文件"
    fi
fi

# 重新加载 systemd
systemctl daemon-reload 2>/dev/null || true

echo "TLCP Channel 安装成功！"
echo "使用 'systemctl start tlcpchan' 启动服务"
echo "使用 'systemctl enable tlcpchan' 设置开机自启"
