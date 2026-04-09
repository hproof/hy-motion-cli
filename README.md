# HY-Motion CLI

命令行工具，用于提交和管理动作生成任务。

## 安装

从源码编译：

```bash
git clone https://gitee.com/hproof/hy-motion-cli.git
cd hy-motion-cli
go build -o hy-motion.exe ./src/main.go
```

## 配置

复制 `config.example.toml` 为 `config.toml`，填入 API 地址和认证信息：

```toml
[api]
url = "http://your-ecs-ip:8000"
timeout = 30

[auth]
user_id = "your-user-id"
token = "your-token"
```

## 使用

```bash
# 提交任务
hy-motion.exe submit "hello world"

# 查看任务状态
hy-motion.exe status <task_id>

# 查看队列状态
hy-motion.exe queue

# 指定配置文件
hy-motion.exe -c config.toml submit "hello"
```
