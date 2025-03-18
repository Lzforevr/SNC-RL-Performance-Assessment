#!/bin/bash
set -e
export GOPATH=/home/user/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export HOME=/home/user
cd /home/user/Future
sudo -u user kubectl rollout status deployment/metrics-server -n kube-system --timeout=1m
sudo -u user kubectl rollout status deployment/kube-state-metrics -n kube-system --timeout=1m
go run promql.go 30