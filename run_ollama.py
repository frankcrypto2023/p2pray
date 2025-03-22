import subprocess
import ray

@ray.remote(num_gpus=1)
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
