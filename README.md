# 🐝 Beehive — 智能体蜂群框架

**让多个 AI Agent 像蜜蜂一样协同工作。**

> 版本: v1.0 | 创建: 2026-05-12
> 详细文档: [架构设计](docs/architecture.md) · [产品需求](docs/PRD.md) · [架构参考图](docs/architecture.svg)

---

## 架构全景

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         主控层 (Brain — 编排)                             │
│  ┌─────────────────────┐                                                │
│  │    🐝 主控服务器      │                                                │
│  │                     │    · DAG 调度与编排                              │
│  │   ┌───────────────┐ │    · Worker 生命周期管理                          │
│  │   │  Coordinator  │ │    · 状态追踪与重试                              │
│  │   │  (beehive)    │ │    · 日志与指标监控                              │
│  │   └───────────────┘ │                                                │
│  └────────┬────────────┘                                                │
│           │ 创建 / 调度 / 监控 Worker 节点                                 │
└───────────┼────────────────────────────────────────────────────────────┘
            │
    ┌───────┼───────────────────────────────────────────────┐
    │       ▼                                               │
    │  ┌──────────────────────────────────────┐             │
    │  │     定义层 (DAG — 任务图)              │  ◄── 双向   │
    │  │                                      │             │
    │  │  · 任务节点与依赖关系                   │             │
    │  │  · 并行分支与汇聚                      │             │
    │  │  · 失败重试策略                        │             │
    │  │  · 资源与配置信息                      │             │
    │  └──────────────────────────────────────┘             │
    └───────────────────────────────────────────────────────┘
            │
            ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                      执行层 (Hands — Worker 节点)                         │
│                                                                          │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐              │
│  │  Worker 1  │  │  Worker 2  │  │  Worker 3  │  │  Worker N  │  ...      │
│  │            │  │            │  │            │  │            │              │
│  │ 🐳 沙箱    │  │ 🐳 沙箱    │  │ 🐳 沙箱    │  │ 🐳 沙箱    │              │
│  │ 🤖 Harness│  │ 🤖 Harness│  │ 🤖 Harness│  │ 🤖 Harness│              │
│  │ 🧠 模型    │  │ 🧠 模型    │  │ 🧠 模型    │  │ 🧠 模型    │              │
│  │ ⚒️ Skill  │  │ ⚒️ Skill  │  │ ⚒️ Skill  │  │ ⚒️ Skill  │              │
│  │ 🕒 Session│  │ 🕒 Session│  │ 🕒 Session│  │ 🕒 Session│              │
│  │ 🔑 SSH Key│  │ 🔑 SSH Key│  │ 🔑 SSH Key│  │ 🔑 SSH Key│              │
│  │ 📁 工作目录│  │ 📁 工作目录│  │ 📁 工作目录│  │ 📁 工作目录│              │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘              │
│        │              │              │              │                     │
│        └──────┬───────┴──────┬───────┘──────┬───────┘                     │
│               │   DAG 依赖    │  并行分支     │                            │
│               ▼              ▼              ▼                            │
│                     🔄 随时销毁 / 重建                                     │
└──────────────────────────────────────────────────────────────────────────┘
```

### Worker 7 组件

| 组件 | 说明 |
|------|------|
| 🐳 沙箱 | Docker 或 CubeSandbox KVM MicroVM |
| 🤖 Agent Harness | 执行框架 (cc-connect agent) |
| 🧠 模型 | Claude Code / Codex / 可配置 |
| ⚒️ Skill | 从 workers/*.yaml 加载 (Dev + Reviewer 成对) |
| 🕒 Session 恢复点 | 从 store/sessions/ 事件日志恢复上下文 |
| 🔑 Git SSH Key | 模板注入，不进沙箱 (安全内置) |
| 📁 工作目录 | 只读挂载项目代码与数据 |

---

## 设计原则

| # | 原则 | 来源 | 含义 |
|---|------|------|------|
| 1 | 接口稳定，实现可变 | Anthropic | Execute 接口不变，实现可替换 |
| 2 | 文件即记忆 | skillssss | 所有产出持久化，不依赖内存上下文 |
| 3 | 隔离即规范 | skillssss | 子 Agent 只接收给定信息 |
| 4 | Cattle, not Pets | Anthropic | Worker 随时销毁/重建，崩溃从 Session 恢复 |
| 5 | 按需供应 | Anthropic | 沙箱仅在 Execute 时创建 |
| 6 | 安全内置 | CubeSandbox | Token 不进入沙箱，KVM + eBPF 隔离 |
| 7 | 纠错闭环 | skillssss | Dev→Review→Fix 最多 3 轮自动修正 |
| 8 | 快照回滚 | CubeSandbox | 修正失败可回滚到修复前状态 |

---

## 执行流水线

```
用户丢 task.yaml → Coordinator 检测 → Decomposer 拆解 DAG → Session 写事件
  → Executor 扫描就绪子任务 →
  ┌─ 单子任务流水线 ──────────────────────────┐
  │ 1. Execute(Dev Agent)  → 加载 Skill → 执行 │
  │ 2. Execute(Reviewer Agent) → 输出审查报告   │
  │ 3. PASS → done  │  FAIL → Fix (最多3轮)    │
  └────────────────────────────────────────────┘
  → lessons-learned 更新 → main-log 记录
```

| 任务类型 | DAG 流程 | 并行 |
|---------|---------|------|
| `feature` | 分析 → 编码(并行) → Review(并行) → 测试(并行) → 合并 | 是 |
| `bugfix` | 诊断 → 修复 → 验证 | 否 |
| `refactor` | 方案 → 重构 → 验证 | 否 |
| `research` | 搜索(并行) → 写作(并行) → 汇总 | 是 |

> 详细执行流程、状态机、数据模型见 [架构设计文档](docs/architecture.md)

---

## 快速开始

```bash
go build -o bin/beehive ./cmd/beehive/

./bin/beehive start      # 终端1: 启动 Coordinator
./bin/beehive worker      # 终端2: Worker 持续循环

# 或通过 Hermes cron 定时调度:
# cc-connect cron add --cron "*/1 * * * *" --exec "cd ~/code/Beehive && ./bin/beehive run" --desc "Beehive Worker"

cp tasks/example.yaml tasks/my-task.yaml   # 丢个任务
./bin/beehive status      # 查看状态
```

---

## CLI 命令

| 命令 | 功能 |
|------|------|
| `beehive start` | 启动 Coordinator，轮询 tasks/ 拆解 DAG |
| `beehive run` | 单次扫描 + 执行一个就绪子任务完整流水线 |
| `beehive worker` | 持续循环模式 |
| `beehive status` | 从 Session 事件日志推导并展示全局状态 |
| `beehive serve` | Web Dashboard (远期) |

---

## 目录结构

```
Beehive/
├── cmd/beehive/main.go       # CLI 入口
├── internal/
│   ├── task/                 # 任务模型 + DAG 拆解
│   ├── session/              # Session 事件日志 (append-only jsonl)
│   ├── executor/             # Execute 接口 + SubAgent/CubeSandbox/Local 实现
│   ├── registry/             # Agent Registry (修正时 resume)
│   ├── review/               # 审查报告模型 (verdict + severity)
│   └── store/                # YAML 读写
├── workers/                  # Worker Skill 定义 (Dev + Reviewer 成对)
├── tasks/                    # 用户丢 .yaml 任务文件
├── store/                    # 运行时: sessions/ agents/ results/ knowledge/
└── docs/                     # 架构设计图 + 详细文档
```

---

## 文档索引

| 文档 | 内容 |
|------|------|
| [docs/PRD.md](docs/PRD.md) | 产品需求: 功能清单(28项)、用户故事、竞品对比、里程碑 |
| [docs/architecture.md](docs/architecture.md) | 架构设计: 数据模型、核心接口、纠错闭环、安全模型、部署 |
| [docs/architecture.svg](docs/architecture.svg) | 架构设计参考图 (三层架构 + Worker 组成 + 示例流程) |

---

## 参考资料

| 来源 | 核心贡献 | 详情 |
|------|---------|------|
| [Anthropic Managed Agents](https://www.anthropic.com/engineering/managed-agents) | Brain-Hands-Session 三元解耦、Cattle 原则、安全边界 | [§17.2](docs/architecture.md) |
| skillssss/harness | 主-从协同、Dev→Review→Fix、文件即记忆 | [§17.4](docs/architecture.md) |
| [CubeSandbox](https://github.com/TencentCloud/CubeSandbox) | KVM MicroVM、<60ms 冷启动、E2B 兼容 | [§17.3](docs/architecture.md) |

---

## 竞品对比

| 维度 | Beehive | LangGraph | CrewAI | AutoGen |
|------|---------|-----------|--------|---------|
| 语言 | Go | Python | Python | Python |
| 执行模型 | DAG + 纠错闭环 | StateGraph | 顺序/层级 | 对话式 |
| 沙箱隔离 | KVM 硬件级 | Docker | 无 | Docker |
| Session | append-only 事件日志 | Checkpointer | 无 | 无 |
| 质量保障 | Dev→Review→Fix 3轮 | 无 | 无 | 无 |
| 部署 | 单二进制 | Python env | Python env | Python env |
