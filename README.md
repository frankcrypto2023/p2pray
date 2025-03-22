# 利用libp2p+Ray+ollama构建共享去中心化GPU 调用大模型
## Ray 是基于python的 GPU集群运算框架 [https://github.com/ray-project/ray](https://github.com/ray-project/ray)
## 1.环境准备
- 客户端环境拥有用 nvidia GPU,CUDA
- ollama 安装 [https://ollama.com/download](https://ollama.com/download)
- `nvidia-smi.exe -l` || `nvidia-smi -l` 验证
```bash
# conda create --name p2pray python=3.12.3
# conda activate p2pray
# python -V
# Python 3.12.3
# pip install 'ray[default]'
# ray --version
# ray, version 2.44.0
```
## 2.启动head节点
```bash
# ray start --head --node-ip-address=172.20.133.120 --port=6379 --dashboard-host=172.20.133.120 --dashboard-port=8265
# 查看状态
# ray status
# 停止ray服务
# ray stop
# 查看控制面板
# http://172.20.133.120:8265
```
## 2.启动cluster worker节点
```bash
# Windows 设置
# $env:RAY_ENABLE_WINDOWS_OR_OSX_CLUSTER=1
# Mac
# export RAY_ENABLE_WINDOWS_OR_OSX_CLUSTER=1
# ray start --address='(head节点ip):6379' --node-ip-address=192.168.0.107 
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
# ray start --address='192.168.0.107:6380' --node-ip-address=192.168.0.107
# ray status 查看状态
```
## 5.提交任务
```bash
# ray job submit --address http://172.20.133.120:8265 --working-dir . -- python task.py
```
## 6.大模型任务
```bash
# ollama pull llama3.2
```
## 7.python 执行ollama命令 run_ollama.py
```python
import subprocess
import ray

@ray.remote(num_gpus=2)
def call_with_ollama(model,input):
    # 构造命令，假设 Ollama 支持传入训练数据和输出路径
    cmd = f'ollama run {model} "{input}"'
    result = subprocess.run(cmd, shell=True, capture_output=True, encoding="utf-8", text=True)
    if result.returncode != 0:
        raise RuntimeError(f"任务 {model} {input} 失败: {result.stderr}")
    return f"任务 {model} {input} 成功: {result.stdout}"

ray.init()

# 假设有4个任务，每个任务处理不同的数据批次
tasks = [call_with_ollama.remote("llama3.2", f"hello {i}") for i in range(4)]
results = ray.get(tasks)
for res in results:
    print(res)
ray.shutdown()

```
## 8.提交大模型任务
```bash
# Ubuntu
# ray job submit --address http://172.20.133.120:8265 --working-dir . -- python run_ollama.py
# Windows
# # ray job submit --address http://172.20.133.120:8265 -- python run_ollama.py
```