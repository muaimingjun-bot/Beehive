package task

import "time"

// Status 子任务状态
type Status string

const (
	StatusPending Status = "pending"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

// SubTask DAG 中的一个子任务节点
type SubTask struct {
	ID          string   `yaml:"id" json:"id"`
	ParentID    string   `yaml:"parent_id" json:"parent_id"`
	Seq         int      `yaml:"seq" json:"seq"`
	Type        string   `yaml:"type" json:"type"`               // code / review / research / test
	Worker      string   `yaml:"worker" json:"worker"`           // 对应 workers/*.yaml
	Prompt      string   `yaml:"prompt" json:"prompt"`
	DependsOn   []string `yaml:"depends_on" json:"depends_on"`   // 前置子任务 ID 列表
	Status      Status   `yaml:"status" json:"status"`
	Context     Context  `yaml:"context" json:"context"`
	ResultFile  string   `yaml:"result_file,omitempty" json:"result_file,omitempty"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	StartedAt   time.Time `yaml:"started_at,omitempty" json:"started_at,omitempty"`
	FinishedAt  time.Time `yaml:"finished_at,omitempty" json:"finished_at,omitempty"`
	Error       string   `yaml:"error,omitempty" json:"error,omitempty"`
}

// Context 子任务共享的上下文
type Context struct {
	TaskDescription string            `yaml:"task_description" json:"task_description"`
	ProjectPath     string            `yaml:"project_path,omitempty" json:"project_path,omitempty"`
	Branch          string            `yaml:"branch,omitempty" json:"branch,omitempty"`
	Extra           map[string]string `yaml:"extra,omitempty" json:"extra,omitempty"`
}

// Task 用户提交的顶层任务
type Task struct {
	ID          string    `yaml:"id" json:"id"`
	Name        string    `yaml:"name" json:"name"`
	Type        string    `yaml:"type" json:"type"` // feature / bugfix / refactor / research
	Description string    `yaml:"description" json:"description"`
	Workers     []string  `yaml:"workers" json:"workers"`
	ProjectPath string    `yaml:"project_path,omitempty" json:"project_path,omitempty"`
	Branch      string    `yaml:"branch,omitempty" json:"branch,omitempty"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	SubTaskIDs  []string  `yaml:"subtask_ids,omitempty" json:"subtask_ids,omitempty"`
	Status      Status    `yaml:"status" json:"status"`
}
