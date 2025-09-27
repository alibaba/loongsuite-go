// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package task

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var defaultStorage *Storage

type Storage struct {
	filePath string
	mu       sync.RWMutex
	cache    map[string]*Task
	dirty    bool
}

func GetDefaultStorage() *Storage {
	if defaultStorage == nil {
		InitDefaultStorage("./data/task")
	}
	return defaultStorage
}

func InitDefaultStorage(dataDir string) error {
	s, err := NewStorage(dataDir)
	if err != nil {
		return err
	}
	defaultStorage = s
	return nil
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}
	s := &Storage{
		filePath: filepath.Join(dataDir, "tasks.jsonl"),
		cache:    make(map[string]*Task),
	}

	if err := s.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load from disk: %v", err)
	}

	return s, nil
}

func (s *Storage) loadFromDisk() error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var task Task
		if err := json.Unmarshal(scanner.Bytes(), &task); err != nil {
			return fmt.Errorf("failed to unmarshal task: %v", err)
		}
		s.cache[task.ID] = &task
	}

	return scanner.Err()
}

func (s *Storage) Add(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.CreatedAt = time.Now().Format(time.RFC3339)
	task.IsDeleted = false
	s.cache[task.ID] = task

	// 直接追加到文件末尾
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %v", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write task: %v", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %v", err)
	}

	return nil
}

func (s *Storage) List(params *ListParams) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var activeTasks, completedTasks []*Task
	for _, task := range s.cache {
		if task.IsDeleted {
			continue
		}

		if params.Query != "" && !contains(task.Title, params.Query) && !contains(task.Content, params.Query) {
			continue
		}

		if params.IsDone != nil {
			if task.Completed != *params.IsDone {
				continue
			}
		}

		if task.Completed {
			completedTasks = append(completedTasks, task)
		} else {
			activeTasks = append(activeTasks, task)
		}
	}

	// 按创建时间排序（最新的在前面）
	sort.Slice(activeTasks, func(i, j int) bool {
		return activeTasks[i].CreatedAt > activeTasks[j].CreatedAt
	})
	sort.Slice(completedTasks, func(i, j int) bool {
		return completedTasks[i].CreatedAt > completedTasks[j].CreatedAt
	})

	// 合并列表：未完成的在前，已完成的在后
	tasks := append(activeTasks, completedTasks...)

	if params.Limit != nil && len(tasks) > *params.Limit {
		tasks = tasks[:*params.Limit]
	}

	return tasks, nil
}

func (s *Storage) Update(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.cache[task.ID]
	if !exists || existing.IsDeleted {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// 只更新非空字段
	updated := *existing // 创建副本
	if task.Title != "" {
		updated.Title = task.Title
	}
	if task.Content != "" {
		updated.Content = task.Content
	}
	if task.Deadline != "" {
		updated.Deadline = task.Deadline
	}
	// Completed 字段需要特殊处理，因为它是布尔值
	if task.Completed != existing.Completed {
		updated.Completed = task.Completed
	}

	s.cache[task.ID] = &updated
	s.dirty = true

	return s.syncToDisk()
}

func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.cache[id]
	if !exists || task.IsDeleted {
		return fmt.Errorf("task not found: %s", id)
	}

	// 标记删除
	task.IsDeleted = true
	s.dirty = true

	return s.syncToDisk()
}

func (s *Storage) syncToDisk() error {
	if !s.dirty {
		return nil
	}

	// 创建临时文件
	tmpFile := s.filePath + ".tmp"
	file, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer file.Close()

	// 写入数据到临时文件
	for _, task := range s.cache {
		data, err := json.Marshal(task)
		if err != nil {
			os.Remove(tmpFile) // 清理临时文件
			return fmt.Errorf("failed to marshal task: %v", err)
		}

		if _, err := file.Write(append(data, '\n')); err != nil {
			os.Remove(tmpFile) // 清理临时文件
			return fmt.Errorf("failed to write task: %v", err)
		}
	}

	// 确保所有数据都写入磁盘
	if err := file.Sync(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to sync file: %v", err)
	}

	// 关闭文件
	if err := file.Close(); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to close file: %v", err)
	}

	// 备份现有文件（如果存在）
	if _, err := os.Stat(s.filePath); err == nil {
		backupFile := s.filePath + ".bak"
		if err := os.Rename(s.filePath, backupFile); err != nil {
			os.Remove(tmpFile)
			return fmt.Errorf("failed to backup file: %v", err)
		}
	}

	// 将临时文件重命名为正式文件
	if err := os.Rename(tmpFile, s.filePath); err != nil {
		// 如果重命名失败，尝试恢复备份
		if backupErr := os.Rename(s.filePath+".bak", s.filePath); backupErr != nil {
			return fmt.Errorf("failed to rename temp file and restore backup: %v, backup error: %v", err, backupErr)
		}
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	// 删除备份文件
	os.Remove(s.filePath + ".bak")

	s.dirty = false
	return nil
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
