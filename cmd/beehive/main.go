package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/muaimingjun/beehive/internal/dispatcher"
	"github.com/muaimingjun/beehive/internal/store"
	"github.com/muaimingjun/beehive/internal/task"
)

func main() {
	baseDir := getBaseDir()
	storeDir := filepath.Join(baseDir, "store")
	tasksDir := filepath.Join(baseDir, "tasks")
	ensureDirs(storeDir)

	cmd := "start"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "status":
		showStatus(storeDir)
	case "ready":
		showReady(storeDir) // 打印第一个可执行子任务的详情
	default:
		runCoordinator(tasksDir, storeDir)
	}
}

func runCoordinator(tasksDir, storeDir string) {
	d := &dispatcher.Dispatcher{StoreDir: storeDir}
	dec := &task.Decomposer{}

	log.Println("🐝 Beehive Coordinator started")
	log.Printf("   Tasks dir: %s", tasksDir)
	log.Printf("   Store dir:  %s", storeDir)
	log.Println("   Drop .yaml tasks into tasks/ — I'll decompose them.")

	for {
		entries, err := os.ReadDir(tasksDir)
		if err != nil {
			log.Printf("read tasks dir: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
				continue
			}
			processTaskFile(filepath.Join(tasksDir, entry.Name()), d, dec)
		}

		time.Sleep(3 * time.Second)
	}
}

func processTaskFile(path string, d *dispatcher.Dispatcher, dec *task.Decomposer) {
	var t task.Task
	if err := store.UnmarshalYAML(path, &t); err != nil {
		log.Printf("parse task %s: %v", filepath.Base(path), err)
		return
	}

	if t.ID != "" {
		return // 已拆解
	}

	log.Printf("📋 New task: %s (type=%s)", t.Name, t.Type)

	subs := dec.Decompose(&t)
	log.Printf("   → %d subtasks generated", len(subs))

	if err := store.WriteFile(path, t); err != nil {
		log.Printf("save task: %v", err)
		return
	}

	ids, err := d.DispatchAll(subs)
	if err != nil {
		log.Printf("dispatch subtasks: %v", err)
		return
	}

	for i, id := range ids {
		log.Printf("   📝 %s (%s)", id, subs[i].Type)
	}
}

func showStatus(storeDir string) {
	counts := map[string]int{}
	var ready, blocked int

	for _, phase := range []string{"pending", "running", "done"} {
		dir := filepath.Join(storeDir, "tasks", phase)
		entries, _ := os.ReadDir(dir)
		counts[phase] = len(entries)
	}

	// 检查 pending 中哪些可执行
	pendDir := filepath.Join(storeDir, "tasks", "pending")
	entries, _ := os.ReadDir(pendDir)
	doneDir := filepath.Join(storeDir, "tasks", "done")
	for _, e := range entries {
		var sub task.SubTask
		if err := store.UnmarshalYAML(filepath.Join(pendDir, e.Name()), &sub); err != nil {
			continue
		}
		if allDepsDone(sub.DependsOn, doneDir) {
			ready++
		} else {
			blocked++
		}
	}

	fmt.Println("🐝 Beehive Status")
	fmt.Println("")
	fmt.Printf("  pending:  %d  (%d ready, %d blocked)\n", counts["pending"], ready, blocked)
	fmt.Printf("  running:  %d\n", counts["running"])
	fmt.Printf("  done:     %d\n", counts["done"])
	fmt.Printf("  failed:   %d\n", countByStatus(filepath.Join(storeDir, "tasks", "running"), task.StatusFailed))
	fmt.Println("")
	fmt.Printf("  Next ready: beehive ready\n")

	// 最近完成的
	if counts["done"] > 0 {
		fmt.Println("\n  Recent done:")
		done := listRecent(filepath.Join(storeDir, "tasks", "done"), 5)
		for _, s := range done {
			fmt.Printf("    ✅ %s (%s)\n", s.ID, s.Type)
		}
	}
}

func showReady(storeDir string) {
	pendDir := filepath.Join(storeDir, "tasks", "pending")
	doneDir := filepath.Join(storeDir, "tasks", "done")
	entries, _ := os.ReadDir(pendDir)

	for _, e := range entries {
		var sub task.SubTask
		if err := store.UnmarshalYAML(filepath.Join(pendDir, e.Name()), &sub); err != nil {
			continue
		}
		if allDepsDone(sub.DependsOn, doneDir) {
			data, _ := store.MarshalYAML(sub)
			fmt.Print(string(data))
			return
		}
	}

	fmt.Println("(no ready subtask)")
}

func allDepsDone(deps []string, doneDir string) bool {
	for _, dep := range deps {
		if _, err := os.Stat(filepath.Join(doneDir, dep+".yaml")); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func countByStatus(dir string, status task.Status) int {
	entries, _ := os.ReadDir(dir)
	count := 0
	for _, e := range entries {
		var sub task.SubTask
		if err := store.UnmarshalYAML(filepath.Join(dir, e.Name()), &sub); err != nil {
			continue
		}
		if sub.Status == status {
			count++
		}
	}
	return count
}

func listRecent(dir string, n int) []task.SubTask {
	entries, _ := os.ReadDir(dir)
	var subs []task.SubTask
	for i := len(entries) - 1; i >= 0 && len(subs) < n; i-- {
		var sub task.SubTask
		if store.UnmarshalYAML(filepath.Join(dir, entries[i].Name()), &sub) == nil {
			subs = append(subs, sub)
		}
	}
	return subs
}

func ensureDirs(storeDir string) {
	dirs := []string{
		filepath.Join(storeDir, "tasks", "pending"),
		filepath.Join(storeDir, "tasks", "running"),
		filepath.Join(storeDir, "tasks", "done"),
		filepath.Join(storeDir, "sessions"),
		filepath.Join(storeDir, "results"),
	}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}
}

func getBaseDir() string {
	if dir := os.Getenv("BEEHIVE_DIR"); dir != "" {
		return dir
	}
	return "."
}

func init() {
	log.SetFlags(0)
	log.SetPrefix("")
}
