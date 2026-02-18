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
ln -sf /etc/tlcpchan/tlcpchan-ui /usr/bin/tlcpchan-ui
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

# 重新加载 systemd
systemctl daemon-reload 2>/dev/null || true

echo "TLCP Channel 安装成功！"
echo "使用 'systemctl start tlcpchan' 启动服务"
echo "使用 'systemctl enable tlcpchan' 设置开机自启"
