#!/bin/bash
# TLCP Channel 安装脚本

set -e

# 创建用户和组
getent group tlcpchan >/dev/null || groupadd -r tlcpchan
getent passwd tlcpchan >/dev/null || useradd -r -g tlcpchan -d /opt/tlcpchan -s /sbin/nologin tlcpchan

# 创建目录
mkdir -p /opt/tlcpchan/{certs/tlcp,certs/tls,logs,ui}
mkdir -p /etc/tlcpchan
mkdir -p /var/log/tlcpchan

# 复制文件
cp tlcpchan /usr/bin/
cp config/config.yaml /etc/tlcpchan/
cp -r certs/* /opt/tlcpchan/certs/ 2>/dev/null || true

# 设置权限
chown -R tlcpchan:tlcpchan /opt/tlcpchan
chown -R tlcpchan:tlcpchan /var/log/tlcpchan
chmod 600 /opt/tlcpchan/certs/tlcp/*.key 2>/dev/null || true
chmod 600 /opt/tlcpchan/certs/tls/*.key 2>/dev/null || true

# 安装systemd服务
cp release/systemd/tlcpchan.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable tlcpchan

echo "安装完成！"
echo "配置文件: /etc/tlcpchan/config.yaml"
echo "启动服务: systemctl start tlcpchan"
