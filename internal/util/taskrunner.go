// Copyright 2023 Harness, Inc.
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

package util

import (
	"context"
	"fmt"
	"sync"
)

// Task represents a task to be executed.
type Task struct {
	ID      int
	Execute func(ctx context.Context) (any, error)
}

// Worker represents a worker that executes tasks.
type Worker struct {
	ID       int
	TaskCh   chan *Task
	ResultCh chan *Result
	Wg       *sync.WaitGroup
}

// NewWorker creates a new worker.
func NewWorker(id int, taskCh chan *Task, resultCh chan *Result, Wg *sync.WaitGroup) *Worker {
	return &Worker{
		ID:       id,
		TaskCh:   taskCh,
		ResultCh: resultCh,
		Wg:       Wg,
	}
}

// Result represents the result of a task.
type Result struct {
	ID   int
	Data any
	Err  error
}

// TaskPool represents a pool of workers that can execute tasks concurrently.
type TaskPool struct {
	MaxParallelism int
	TaskCh         chan *Task
	ResultCh       chan *Result
	Wg             *sync.WaitGroup
	ctx            context.Context
	Cancel         context.CancelFunc
}

// Start starts the thread pool with the specified number of workers.
func (tp *TaskPool) Start() error {
	for i := 0; i < tp.MaxParallelism; i++ {
		worker := NewWorker(i, tp.TaskCh, tp.ResultCh, tp.Wg)
		go func() {
			err := worker.Start(tp.ctx)
			if err != nil {
				fmt.Printf("failed to start a worker: %v", err)
			}
		}()
	}
	return nil
}

// NewTaskPool creates a new job pool.
func NewTaskPool(ctx context.Context, maxParallelism int) *TaskPool {
	taskCh := make(chan *Task)
	resultCh := make(chan *Result)
	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}

	return &TaskPool{
		MaxParallelism: maxParallelism,
		TaskCh:         taskCh,
		ResultCh:       resultCh,
		ctx:            ctx,
		Cancel:         cancel,
		Wg:             wg,
	}
}

// Start starts the worker to listen for tasks and execute them.
func (w *Worker) Start(ctx context.Context) error {
	for {
		select {
		case task, ok := <-w.TaskCh:
			if !ok {
				return nil // Task channel closed
			}
			result, err := task.Execute(ctx)
			w.ResultCh <- &Result{ID: task.ID, Data: result, Err: err}
		case <-ctx.Done():
			return ctx.Err() // Context canceled
		}
	}
}

func (tp *TaskPool) MarkResultRead() {
	tp.Wg.Done()
}

// Submit submits a task to the thread pool.
func (tp *TaskPool) Submit(task *Task) {
	tp.Wg.Add(1)
	tp.TaskCh <- task
}

// Shutdown gracefully shuts down the thread pool
func (tp *TaskPool) Shutdown() {
	tp.Wg.Wait()     // Wait for all tasks to complete
	close(tp.TaskCh) // Signal workers to stop accepting new tasks
	tp.Cancel()      // Cancel the context to stop all ongoing tasks
	close(tp.ResultCh)
}

func (tp *TaskPool) ForceShutdown() {
	tp.Cancel()
	close(tp.TaskCh)
	close(tp.ResultCh)
}
