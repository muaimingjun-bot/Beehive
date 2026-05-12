package dispatcher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/muaimingjun/beehive/internal/store"
	"github.com/muaimingjun/beehive/internal/task"
)

// Dispatcher 将子任务写入文件系统，供 Hermes cron / delegate_task 拿取执行
type Dispatcher struct {
	StoreDir string
}

// Dispatch 将子任务写入 store/tasks/pending/
func (d *Dispatcher) Dispatch(sub task.SubTask) error {
	pendingDir := filepath.Join(d.StoreDir, "tasks", "pending")
	filename := fmt.Sprintf("%s.yaml", sub.ID)
	path := filepath.Join(pendingDir, filename)

	return store.WriteFile(path, sub)
}

// DispatchAll 将一批子任务写入 pending
func (d *Dispatcher) DispatchAll(subs []task.SubTask) ([]string, error) {
	var ids []string
	for _, sub := range subs {
		if err := d.Dispatch(sub); err != nil {
			return ids, err
		}
		ids = append(ids, sub.ID)
	}
	return ids, nil
}

// MarkRunning 将子任务从 pending 移到 running
func (d *Dispatcher) MarkRunning(subID string) error {
	return move(
		filepath.Join(d.StoreDir, "tasks", "pending", subID+".yaml"),
		filepath.Join(d.StoreDir, "tasks", "running", subID+".yaml"),
	)
}

// MarkDone 将子任务从 running 移到 done
func (d *Dispatcher) MarkDone(subID string) error {
	return move(
		filepath.Join(d.StoreDir, "tasks", "running", subID+".yaml"),
		filepath.Join(d.StoreDir, "tasks", "done", subID+".yaml"),
	)
}

// MarkFailed 标记失败——把 running 中的文件状态改成 failed 并记录错误
func (d *Dispatcher) MarkFailed(subID string, errMsg string) error {
	path := filepath.Join(d.StoreDir, "tasks", "running", subID+".yaml")

	var sub task.SubTask
	if err := store.UnmarshalYAML(path, &sub); err != nil {
		return fmt.Errorf("read subtask for mark-failed: %w", err)
	}
	sub.Status = task.StatusFailed
	sub.Error = errMsg
	return store.WriteFile(path, sub)
}

// UpdateStatus 通用状态更新
func (d *Dispatcher) UpdateStatus(subID string, status task.Status, errMsg string) error {
	path := filepath.Join(d.StoreDir, "tasks", "running", subID+".yaml")

	var sub task.SubTask
	if err := store.UnmarshalYAML(path, &sub); err != nil {
		return fmt.Errorf("read subtask for update: %w", err)
	}
	sub.Status = status
	sub.Error = errMsg
	return store.WriteFile(path, sub)
}

func move(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("move %s → %s: %w", src, dst, err)
	}
	return nil
}
