# HY-Motion CLI

命令行工具，用于提交和管理动作生成任务。

## 安装

从源码编译：

```bash
git clone https://gitee.com/hproof/hy-motion-cli.git
cd hy-motion-cli
go build -o hy-motion-cli.exe ./src
```

## 配置

配置文件固定为 `.hy-motion-cli.toml`，程序会自动查找：
1. 先从当前目录查找
2. 当前目录不存在则从 home 目录查找

复制 `.hy-motion-cli.toml.example` 为 `.hy-motion-cli.toml`（或直接创建），填入 API 地址和认证信息：

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
hy-motion-cli submit "hello world"

# 查看任务状态
hy-motion-cli status <task_id>

# 查看队列状态
hy-motion-cli queue
```
