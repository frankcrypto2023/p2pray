import ray

# 初始化 Ray
ray.init(address='auto')  # 自动连接到本地或集群中的 Ray 实例

# 定义一个简单的远程函数
@ray.remote
def simple_task():
    return "Hello, Ray!"

# 提交任务并获取结果
future = simple_task.remote()
result = ray.get(future)
print(result)  # 输出: Hello, Ray!

# 关闭 Ray
ray.shutdown()

