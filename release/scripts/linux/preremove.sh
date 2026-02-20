#!/bin/bash
set -e

if systemctl is-active --quiet tlcpchan 2>/dev/null; then
    systemctl stop tlcpchan
fi

if systemctl is-enabled --quiet tlcpchan 2>/dev/null; then
    systemctl disable tlcpchan
fi

systemctl daemon-reload 2>/dev/null || true

# 删除软链接
rm -f /usr/bin/tlcpchan
rm -f /usr/bin/tlcpchan-cli
rm -f /usr/bin/tlcpc
