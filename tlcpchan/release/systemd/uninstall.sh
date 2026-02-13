#!/bin/bash
# TLCP Channel 卸载脚本

systemctl stop tlcpchan 2>/dev/null || true
systemctl disable tlcpchan 2>/dev/null || true
rm -f /etc/systemd/system/tlcpchan.service
systemctl daemon-reload

rm -f /usr/bin/tlcpchan
rm -rf /etc/tlcpchan
# 保留数据和日志
# rm -rf /opt/tlcpchan

echo "卸载完成！"
