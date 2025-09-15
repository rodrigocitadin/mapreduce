package scheduler

import "github.com/rodrigocitadin/mapreduce/pkg/types"

type TaskQueue struct {
	queue chan *types.Task
}

func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		queue: make(chan *types.Task, 1024),
	}
}

func (tq *TaskQueue) Push(task *types.Task) {
	tq.queue <- task
}

func (tq *TaskQueue) Pop() *types.Task {
	return <-tq.queue
}
