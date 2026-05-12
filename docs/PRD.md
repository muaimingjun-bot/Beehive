# Beehive — 产品需求文档 (PRD)

> 版本: v1.0 | 状态: Draft | 创建: 2026-05-12

---

## 1. 产品概述

### 1.1 定位

Beehive 是一个**智能体蜂群框架**。让多个 AI Agent 像蜜蜂一样协同工作——用户丢入任务描述，系统自动拆解为 DAG 子任务，编排多个子 Agent 按依赖关系执行，经过 Dev→Review→Fix 纠错闭环后产出高质量结果。

### 1.2 核心价值

| 痛点 | Beehive 解法 |
|------|-------------|
| 复杂任务需手动拆解分配 | **自动 DAG 拆解** — 按任务类型生成子任务依赖图 |
| 单 Agent 长对话上下文爆炸 | **隔离执行** — 每个子任务独立 Agent |
| Agent 输出质量不可控 | **纠错闭环** — Dev→Review→Fix 最多 3 轮 |
| 多任务无法并行 | **并行 Worker** — 无依赖节点并发执行 |
| 执行过程不可追溯 | **Session 事件日志** — append-only 全流程记录 |
| LLM 代码安全风险 | **CubeSandbox KVM 隔离** — 内核级隔离 |

### 1.3 目标用户

- **AI 辅助开发团队** — 自动化 分析→编码→审查→测试
- **研究者** — 并行搜索多主题，自动汇总报告
- **Agent 编排开发者** — YAML 模板自定义 Worker 和执行流程

---

## 2. 功能需求

### 2.1 任务提交与管理

| ID | 功能 | 优先级 | 描述 |
|----|------|--------|------|
| F1 | 任务 YAML 提交 | P0 | 放入 tasks/ 目录即可触发 |
| F2 | 任务类型识别 | P0 | feature/bugfix/refactor/research |
| F3 | 自定义任务类型 | P1 | 扩展 decomposer |
| F4 | 任务状态查询 | P0 | `beehive status` |

### 2.2 DAG 拆解与调度

| ID | 功能 | 优先级 | 描述 |
|----|------|--------|------|
| F5 | 自动 DAG 拆解 | P0 | 按类型生成子任务依赖图 |
| F6 | 依赖感知调度 | P0 | 前置完成后才执行后续 |
| F7 | 并行分支执行 | P1 | 无依赖节点并发 |
| F8 | 分支汇聚 | P1 | 并行结果汇总合并 |

### 2.3 Worker 执行引擎

| ID | 功能 | 优先级 | 描述 |
|----|------|--------|------|
| F9 | Execute 接口 | P0 | `Execute(ctx, worker, prompt) → Result` |
| F10 | SubAgent 执行 | P0 | cc-connect Agent |
| F11 | Worker Skill 加载 | P0 | workers/*.yaml |
| F12 | CubeSandbox 执行 | P2 | KVM MicroVM |
| F13 | 本地执行 | P2 | 简单任务 |

### 2.4 纠错闭环

| ID | 功能 | 优先级 | 描述 |
|----|------|--------|------|
| F14 | Dev→Review 流水线 | P0 | 开发 + 审查 |
| F15 | 结构化审查报告 | P0 | test-report.json |
| F16 | 自动修正循环 | P1 | max 3轮 |
| F17 | 严重级别分级 | P1 | blocker/major/minor |
| F18 | Agent Registry | P1 | resume 同一 Agent |

### 2.5 Session 与状态管理

| ID | 功能 | 优先级 | 描述 |
|----|------|--------|------|
| F19 | Session 事件日志 | P0 | append-only jsonl |
| F20 | 状态推导 | P0 | 从事件推导 |
| F21 | 崩溃恢复 | P1 | Session 外置恢复 |
| F22 | 失败重试 | P1 | 可配置次数 |
| F23 | 超时处理 | P1 | 自动标记 |

### 2.6 经验积累与可观测性

| ID | 功能 | 优先级 | 描述 |
|----|------|--------|------|
| F24 | lessons-learned | P2 | 跨任务知识 |
| F25 | main-log | P1 | 全流程日志 |
| F26 | 会话记忆 | P2 | 跨任务上下文 |
| F27 | Web Dashboard | P2 | `beehive serve` |
| F28 | 成本追踪 | P1 | token/迭代统计 |

---

## 3. 非功能需求

| 类别 | 需求 | 目标 |
|------|------|------|
| 性能 | 任务拆解延迟 | < 1s |
| 性能 | Worker 冷启动 | < 60ms (CubeSandbox) |
| 性能 | 单机密度 | 1000+ (< 5MB/实例) |
| 安全 | 沙箱隔离 | KVM 硬件级 + eBPF |
| 安全 | 凭据保护 | Token 不进沙箱 |
| 可靠 | 崩溃恢复 | Session 外置 |
| 可靠 | 数据持久化 | store/results/ |
| 可扩展 | 自定义 Worker | YAML 模板 |
| 可扩展 | 多执行后端 | Execute 接口多实现 |

---

## 4. 用户故事

### US-1: 简单任务

```bash
cp tasks/example-research.yaml tasks/my-task.yaml
./bin/beehive start && ./bin/beehive run
cat store/results/xxx-report.md
```

### US-2: 软件研发全流程

```
feature task.yaml
  → analyze → code(A/B/C 并行) → review(A/B 并行) → test → merge
  → 每步经过 Dev→Review→Fix
```

### US-3: 崩溃恢复

```bash
kill -9 $(pgrep beehive)
./bin/beehive start && ./bin/beehive worker  # 自动恢复
```

---

## 5. 竞品对比

| 维度 | Beehive | LangGraph | CrewAI | AutoGen |
|------|---------|-----------|--------|---------|
| 语言 | Go | Python | Python | Python |
| 执行模型 | DAG + 纠错闭环 | StateGraph | 顺序/层级 | 对话式 |
| 沙箱隔离 | KVM 硬件级 | Docker | 无 | Docker |
| Session | append-only 事件日志 | Checkpointer | 无 | 无 |
| 质量保障 | Dev→Review→Fix 3轮 | 无内置 | 无 | 无 |
| 崩溃恢复 | Session 外置 | Checkpointer | 无 | 无 |
| 部署 | 单二进制 | Python env | Python env | Python env |

---

## 6. 技术选型

| 组件 | 技术 | 理由 |
|------|------|------|
| 主控服务器 | Go 1.25+ | 高性能、单二进制、并发原生 |
| 配置格式 | YAML | 可读性好 |
| 事件日志 | JSONL (append-only) | 可查询、可恢复 |
| 沙箱 | CubeSandbox (KVM) | 内核级隔离、<60ms、E2B兼容 |
| Agent 调用 | cc-connect Agent | 利用现有基础设施 |
| Web Dashboard | Go net/http + embed | 零外部依赖 |

---

## 7. 里程碑

| 阶段 | 内容 | 交付物 |
|------|------|--------|
| **M1** | Session + Execute + SubAgentExecutor | Dev→Review 流水线可运行 |
| **M2** | Agent Registry + 3轮修正 | 自动纠错闭环 |
| **M3** | 失败重试 + 超时 + 日志 | 生产就绪 |
| **M4** | 会话记忆 + Web Dashboard | 可视化 + 跨任务上下文 |
| **M5** | CubeSandbox + 并行分支 | 安全隔离 + 性能 |

---

## 8. 快速开始

```bash
go build -o bin/beehive ./cmd/beehive/
./bin/beehive start      # 终端1: Coordinator
./bin/beehive worker      # 终端2: Worker 循环
cp tasks/example.yaml tasks/my-task.yaml   # 丢任务
./bin/beehive status      # 查看状态
```
