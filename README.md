# 🐝 Beehive — 智能体蜂群框架

**让多个 AI Agent 像蜜蜂一样协同工作。**

## 设计

```
用户丢 task.yaml → Coordinator 检测 → 拆成子任务 DAG → 写文件系统
                                                          ↓
                                            Hermes cron/delegate_task
                                            拿取子任务 → Worker 执行 → 输出结果
```

## 目录

```
Beehive/
├── cmd/beehive/main.go    # 入口，轮询 tasks/
├── internal/
│   ├── task/              # 任务模型 + 拆解器
│   ├── dispatcher/        # 文件系统派发
│   └── store/             # YAML 读写
├── workers/               # Worker 模板（YAML）
├── tasks/                 # 用户丢任务进来
└── store/                 # 运行时状态
    ├── tasks/{pending,running,done}/
    ├── sessions/          # 跨会话上下文
    └── results/           # 产出物
```

## 快速开始

```bash
# 编译
cd ~/code/Beehive
go build -o bin/beehive ./cmd/beehive/

# 启动 Coordinator（持续运行）
./bin/beehive

# 丢个任务试试
cp tasks/example.yaml tasks/my-task.yaml
```

## 任务类型

| type | 流程 | 子任务数 |
|------|------|---------|
| `feature` | 分析 → 实现 → 审查 | 3 |
| `bugfix` | 诊断 → 修复 → 验证 | 3 |
| `refactor` | 方案 → 重构 → 验证 | 3 |
| `research` | 调研 → 分析 | 2 |
| 其他 | 执行 → 检查 | 2 |

## 与 Hermes 集成

Coordinator 只负责拆解和状态管理，执行交给 Hermes：

```bash
# Hermes 定时扫描 pending 子任务并执行（方案待定）
hermes cron "beehive-worker" --schedule "every 30s" \
  --prompt "扫描 ~/code/Beehive/store/tasks/pending/，拿一个无反依赖的，用对应 worker 执行"
```

## 下一步

- [ ] Hermes cron 集成（自动拿取子任务执行）
- [ ] 会话记忆持久化（store/sessions/）
- [ ] CubeSandbox 隔离执行
- [ ] 失败重试和超时处理
- [ ] Web dashboard
