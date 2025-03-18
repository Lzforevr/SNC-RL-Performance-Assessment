#!/bin/bash
set -e
set -x
DIR="/home/user/Future"
USER_PATH="/home/user"

# 镜像导入
cd $DIR
sudo ctr -n k8s.io image import kube-state-metrics.tar.gz --platform=linux/arm64
sudo ctr -n k8s.io image import metrics-server.tar --platform=linux/arm64

# Prometheus 外部部署
sudo mv $DIR/prometheus /usr/local/prometheus
sudo mv $DIR/prometheus.service /usr/lib/systemd/system
sudo systemctl daemon-reload
sudo systemctl start prometheus
sudo systemctl enable prometheus --now

sudo mv $DIR/node_exporter /usr/local/node_exporter
sudo mv $DIR/node_exporter.service /etc/systemd/system/node_exporter.service
sudo systemctl daemon-reload
sudo systemctl start node_exporter
sudo systemctl enable node_exporter --now

# metrics-server 部署
sudo -u user kubectl apply -f metrics-server.yaml
sudo -u user kubectl rollout status deployment/metrics-server -n kube-system --timeout=1m

# kube-state-metrics 部署
cd kube-state-metrics
sudo -u user kubectl apply -f .
sudo -u user kubectl rollout status deployment/kube-state-metrics -n kube-system --timeout=1m

# cAdvisor暴露给Prometheus
sudo -u user kubectl create ns cadvisor
sudo -u user kubectl create serviceaccount monitor -n cadvisor
sudo -u user kubectl create clusterrolebinding monitor-clusterrolebinding -n cadvisor --clusterrole=cluster-admin --serviceaccount=cadvisor:monitor
sudo -u user kubectl create token monitor -n cadvisor --duration=8760h > $USER_PATH/monitor-token

# 设置go脚本开机自启
sudo mv $DIR/future-promql.service /usr/lib/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable --now future-promql
sudo systemctl start future-promql