package main

import (
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
	tasksDir := filepath.Join(baseDir, "tasks")
	storeDir := filepath.Join(baseDir, "store")

	// 确保目录结构存在
	ensureDirs(storeDir)

	d := &dispatcher.Dispatcher{StoreDir: storeDir}
	dec := &task.Decomposer{}

	log.Println("🐝 Beehive Coordinator started")
	log.Printf("   Tasks dir: %s", tasksDir)
	log.Printf("   Store dir:  %s", storeDir)

	// 主循环：轮询 tasks/ 目录
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
			processTaskFile(filepath.Join(tasksDir, entry.Name()), tasksDir, d, dec)
		}

		time.Sleep(3 * time.Second)
	}
}

func processTaskFile(path, tasksDir string, d *dispatcher.Dispatcher, dec *task.Decomposer) {
	// 读取顶层任务
	var t task.Task
	if err := store.UnmarshalYAML(path, &t); err != nil {
		log.Printf("parse task %s: %v", filepath.Base(path), err)
		return
	}

	// 已有 ID 表示已拆解过，跳过
	if t.ID != "" {
		return
	}

	log.Printf("📋 New task: %s (type=%s)", t.Name, t.Type)

	// 拆解成子任务
	subs := dec.Decompose(&t)
	log.Printf("   → %d subtasks generated", len(subs))

	// 写回顶层任务（带 ID 和子任务列表）
	if err := store.WriteFile(path, t); err != nil {
		log.Printf("save task: %v", err)
		return
	}

	// 派发所有子任务
	ids, err := d.DispatchAll(subs)
	if err != nil {
		log.Printf("dispatch subtasks: %v", err)
		return
	}

	for i, id := range ids {
		log.Printf("   📝 subtask %d: %s (%s)", i+1, id, subs[i].Type)
	}
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
