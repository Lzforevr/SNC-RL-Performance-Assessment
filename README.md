# SNC-RL-Performance-Assessment

*Tips:This repository is currently only in demo version.*

### 1.Introduction

  This repository provides the traces and codes for a research on Performance Evaluations for Satellite Cloud Native Architecture. The aim of this research is to scientifically select the optimal combination of cloud-native components through the reinforcement learning-based evaluation method, so as to improve the performance and reliability of the satellite computing system, which in turn improves the computational efficiency of cloud-native satellites and provides support for the low-cost evaluation of new cloud-native components in the future.

### 2.Components

#### 2.1 Trace Data

  Path: /data/Future/2024-12-14 
  
  (Note that the data uploaded on 2024-12-14 is a demo version with Kubernetes and Containerd.)

  The Traces for Combinations of several container runtimes and container orchetrators, currently including information in thr following table:

| Attributes            | Unit     | Description                                                                                            | Example           |
| --------------------- | -------- | ------------------------------------------------------------------------------------------------------ | ----------------- |
| time                  | hh:mm:ss | the Coordinated Universal Time with an interval of 10 seconds                                          | 10:45:12          |
| pod_name              | String   | the pod name of node                                                                                   | calico-node-wc5jb |
| container orchetrator | String   | the software to deploy applications at scale by automating the networking and management of containers | Kubernetes        |
| container runtime     | String   | runtimes that facilitate the creation, execution, and management of containers on a host system.       | containerd        |
| cpu_usage             | Core     | the CPU core consumption                                                                               | 0.030105          |
| cpu_usage_request     | %        | the ratio of the CPU usage to the requested CPU allocation                                             | 17.057705         |
| memory_usage          | Byte     | the working memory usage                                                                               | 23830528          |
| memory_usage_request  | %        | the ratio of memory usage to the requested memory allocation                                           | 32.181920         |
| network_speed_in      | Byte/s   | the speed of flow received by pod                                                                      | 287.957571        |
| network_speed_out     | Byte/s   | the speed of flow transimitted by pod                                                                  | 473.759013        |

The columns above are collected through Prometheus API with extra specific containers.

#### 2.2 Codes for Data Collection & Data Processing
Path:/src/promql.go; /data/Future/merge_code.py

The Data Collection and Processing Codes are written in Go&Python Language. The Data Processing Code is used to merge the data collected from different Prometheus instances.

Run codes by referring to deployment.pdf, you can get the traces in /data/2024-12-14.

Futuremore, the tools we used for data collection are listed below:

1. Prometheus v3.0.0: https://prometheus.io/download/
2. Kubernetes v1.28.15: https://kubernetes.io/docs/tasks/tools/install-kubectl/
3. Containerd v1.7.0: https://github.com/containerd/containerd
4. kube-state-metrics v2.14.0: https://github.com/kubernetes/kube-state-metrics
5. metrics-server v0.6.3: https://github.com/kubernetes-sigs/metrics-server
6. cAdvisor: https://github.com/google/cadvisor

#### 2.3 Model & Training Codes
Upcoming...

#### 2.4 Other items
Upcoming...