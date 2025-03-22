# 利用libp2p+Ray构建共享去中心化GPU 调用大模型
## Ray 是基于python的 GPU集群运算框架 [https://github.com/ray-project/ray](https://github.com/ray-project/ray)
## 1.环境准备
- 客户端环境拥有用 nvidia GPU,CUDA
- `nvidia-smi.exe -l` || `nvidia-smi -l` 验证
```bash
# conda create --name p2pray python=3.12.3
# python -V
# Python 3.12.3
# pip install 'ray[default]'
# ray --version
# ray, version 2.44.0
```
## 2.启动head节点
```bash
# ray start --head --node-ip-address=172.20.133.120 --port=6379 --dashboard-host=172.20.133.120--dashboard-port=8265
# 查看状态
# ray status
# 停止ray服务
# ray stop
```
## 2.启动cluster worker节点
```bash
# Windows 设置
# $env:RAY_ENABLE_WINDOWS_OR_OSX_CLUSTER=1
# Mac
# export RAY_ENABLE_WINDOWS_OR_OSX_CLUSTER=1
# ray start --address='(head节点ip):6379' --node-ip-address=192.168.0.107 --dashboard-port=8265
```
## 3.由于本机机器没有固定IP服务，所以需要借助libp2p 做内容转发服务
```bash
# 编译当前项目
# go build
```
## 4.启动rayp2p 代理服务,并重新启动ray节点 3-4两步可以集成至qng服务，当前只是测试
```bash
# 对于head节点只需要启动proxy
#./rayproxy
# 对于worker节点 需要创建与head节点的连接
# .\rayproxy.exe --peer /ip4/172.20.133.120/tcp/42895/p2p/12D3KooWCVb84GmLUpV1uwc6uHKsJxj8JHTwJiNshmTj3pjBeque --localproxy 192.168.0.107:6380
# worker 节点ray重新启动
# ray start --address='192.168.0.107:6380' --node-ip-address=192.168.0.107 --dashboard-port=8265
# ray status 查看状态
```
## 5.提交任务
```bash
# ray job submit --address http://172.20.133.120:8265 --working-dir . -- python task.py
```
