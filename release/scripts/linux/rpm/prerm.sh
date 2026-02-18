#!/bin/bash
set -e

# 停止服务
if systemctl is-active --quiet tlcpchan 2>/dev/null; then
    systemctl stop tlcpchan
fi

# 禁用服务
if systemctl is-enabled --quiet tlcpchan 2>/dev/null; then
    systemctl disable tlcpchan
fi

# 重新加载 systemd
systemctl daemon-reload 2>/dev/null || true

# 删除软链接
rm -f /usr/bin/tlcpchan
rm -f /usr/bin/tlcpchan-cli
rm -f /usr/bin/tlcpchan-ui
rm -f /usr/bin/tlcpc
