package task

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Decomposer 将顶层任务拆解为子任务 DAG
type Decomposer struct{}

// Decompose 根据任务类型拆解
func (d *Decomposer) Decompose(t *Task) []SubTask {
	t.ID = shortID()
	t.CreatedAt = time.Now()

	switch t.Type {
	case "feature":
		return d.decomposeFeature(t)
	case "bugfix":
		return d.decomposeBugfix(t)
	case "refactor":
		return d.decomposeRefactor(t)
	case "research":
		return d.decomposeResearch(t)
	default:
		return d.decomposeGeneric(t)
	}
}

// feature: 分析 → 实现 → 审查
func (d *Decomposer) decomposeFeature(t *Task) []SubTask {
	ctx := Context{
		TaskDescription: t.Description,
		ProjectPath:     t.ProjectPath,
		Branch:          t.Branch,
	}

	s1 := d.sub(t, 1, "analyze", "researcher",
		fmt.Sprintf("分析需求: %s\n\n制定技术方案，列出需要修改的文件和关键接口设计。输出到 store/results/%s-analyze.md", t.Description, t.ID),
		nil, ctx)
	s2 := d.sub(t, 2, "code", "coder",
		fmt.Sprintf("按照技术方案实现: %s\n\n先读 store/results/%s-analyze.md 获取方案，然后写代码。保持简洁，一个 commit 完成。", t.Description, t.ID),
		[]string{s1.ID}, ctx)
	s3 := d.sub(t, 3, "review", "reviewer",
		fmt.Sprintf("审查代码实现，对照 store/results/%s-analyze.md 的方案，检查逻辑正确性、边界条件和代码风格。输出到 store/results/%s-review.md", t.ID, t.ID),
		[]string{s2.ID}, ctx)

	subs := []SubTask{s1, s2, s3}
	t.SubTaskIDs = idsOf(subs)
	return subs
}

func (d *Decomposer) decomposeBugfix(t *Task) []SubTask {
	ctx := Context{
		TaskDescription: t.Description,
		ProjectPath:     t.ProjectPath,
		Branch:          t.Branch,
	}

	s1 := d.sub(t, 1, "diagnose", "researcher",
		fmt.Sprintf("诊断 bug: %s\n\n定位根因，输出诊断报告到 store/results/%s-diagnose.md", t.Description, t.ID),
		nil, ctx)
	s2 := d.sub(t, 2, "fix", "coder",
		fmt.Sprintf("修复 bug: %s\n\n先读 store/results/%s-diagnose.md，修复并确保不引入新问题。", t.Description, t.ID),
		[]string{s1.ID}, ctx)
	s3 := d.sub(t, 3, "verify", "reviewer",
		fmt.Sprintf("验证修复: %s\n\n确认 bug 已修复，相关测试通过。输出到 store/results/%s-verify.md", t.ID, t.ID),
		[]string{s2.ID}, ctx)

	subs := []SubTask{s1, s2, s3}
	t.SubTaskIDs = idsOf(subs)
	return subs
}

func (d *Decomposer) decomposeRefactor(t *Task) []SubTask {
	ctx := Context{
		TaskDescription: t.Description,
		ProjectPath:     t.ProjectPath,
		Branch:          t.Branch,
	}

	s1 := d.sub(t, 1, "plan", "researcher",
		fmt.Sprintf("重构方案: %s\n\n分析现有代码结构，制定重构计划。输出到 store/results/%s-plan.md", t.Description, t.ID),
		nil, ctx)
	s2 := d.sub(t, 2, "refactor", "coder",
		fmt.Sprintf("执行重构: %s\n\n按 store/results/%s-plan.md 的方案重构，保证行为不变。", t.Description, t.ID),
		[]string{s1.ID}, ctx)
	s3 := d.sub(t, 3, "validate", "reviewer",
		fmt.Sprintf("验证重构: %s\n\n跑测试、对比行为，确认无回归。输出到 store/results/%s-validate.md", t.ID, t.ID),
		[]string{s2.ID}, ctx)

	subs := []SubTask{s1, s2, s3}
	t.SubTaskIDs = idsOf(subs)
	return subs
}

func (d *Decomposer) decomposeResearch(t *Task) []SubTask {
	ctx := Context{
		TaskDescription: t.Description,
		ProjectPath:     t.ProjectPath,
	}

	s1 := d.sub(t, 1, "survey", "researcher",
		fmt.Sprintf("调研: %s\n\n广泛搜索资料、论文、开源项目，汇总到 store/results/%s-survey.md", t.Description, t.ID),
		nil, ctx)
	s2 := d.sub(t, 2, "analyze", "researcher",
		fmt.Sprintf("分析: %s\n\n基于 store/results/%s-survey.md 深入分析，形成结论和建议。输出到 store/results/%s-report.md", t.Description, t.ID, t.ID),
		[]string{s1.ID}, ctx)

	subs := []SubTask{s1, s2}
	t.SubTaskIDs = idsOf(subs)
	return subs
}

func (d *Decomposer) decomposeGeneric(t *Task) []SubTask {
	ctx := Context{
		TaskDescription: t.Description,
		ProjectPath:     t.ProjectPath,
		Branch:          t.Branch,
	}

	s1 := d.sub(t, 1, "execute", "coder",
		t.Description,
		nil, ctx)
	s2 := d.sub(t, 2, "check", "reviewer",
		fmt.Sprintf("检查执行结果: %s\n\n输出到 store/results/%s-check.md", t.Description, t.ID),
		[]string{s1.ID}, ctx)

	subs := []SubTask{s1, s2}
	t.SubTaskIDs = idsOf(subs)
	return subs
}

func idsOf(subs []SubTask) []string {
	ids := make([]string, len(subs))
	for i, s := range subs {
		ids[i] = s.ID
	}
	return ids
}

func (d *Decomposer) sub(t *Task, seq int, typ, worker, prompt string, deps []string, ctx Context) SubTask {
	return SubTask{
		ID:        fmt.Sprintf("%s-%d", t.ID, seq),
		ParentID:  t.ID,
		Seq:       seq,
		Type:      typ,
		Worker:    worker,
		Prompt:    prompt,
		DependsOn: deps,
		Status:    StatusPending,
		Context:   ctx,
		CreatedAt: time.Now(),
	}
}

func shortID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
