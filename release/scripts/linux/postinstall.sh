#!/bin/bash
set -e

if ! getent passwd tlcpchan > /dev/null; then
    useradd -r -s /bin/false -d /etc/tlcpchan tlcpchan
fi

chown -R tlcpchan:tlcpchan /etc/tlcpchan/keystores 2>/dev/null || true
chown -R tlcpchan:tlcpchan /etc/tlcpchan/logs 2>/dev/null || true

# 创建软链接到 /usr/bin
ln -sf /etc/tlcpchan/tlcpchan /usr/bin/tlcpchan
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpchan-cli
ln -sf /etc/tlcpchan/tlcpchan-cli /usr/bin/tlcpc

systemctl daemon-reload 2>/dev/null || true

echo "TLCP Channel 安装成功！"
echo "使用 'systemctl start tlcpchan' 启动服务"
echo "使用 'systemctl enable tlcpchan' 设置开机自启"
