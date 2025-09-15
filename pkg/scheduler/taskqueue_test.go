package scheduler

import (
	"testing"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

func TestTaskQueue(t *testing.T) {
	tq := NewTaskQueue()

	task1 := &types.Task{Id: 1, Type: types.MapTaskType}
	task2 := &types.Task{Id: 2, Type: types.MapTaskType}

	tq.Push(task1)
	tq.Push(task2)

	popped1 := tq.Pop()
	if popped1.Id != 1 {
		t.Errorf("Expected task with ID 1, but got %v", popped1)
	}

	popped2 := tq.Pop()
	if popped2.Id != 2 {
		t.Errorf("Expected task with ID 2, but got %v", popped2)
	}
}

