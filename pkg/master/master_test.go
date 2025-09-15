package master

import (
	"testing"
	"time"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

func TestRegisterWorker(t *testing.T) {
	m := NewMaster()
	args := &types.RegisterWorkerArgs{Addr: "localhost:1234"}
	reply := &types.RegisterWorkerReply{}

	err := m.RegisterWorker(args, reply)
	if err != nil {
		t.Fatalf("RegisterWorker failed: %v", err)
	}
	if reply.WorkerId <= 0 {
		t.Errorf("Expected a positive worker ID, but got %d", reply.WorkerId)
	}

	m.mu.Lock()
	if len(m.workers) != 1 {
		t.Errorf("Expected 1 worker to be registered, but found %d", len(m.workers))
	}
	m.mu.Unlock()
}

func TestMasterFullFlow(t *testing.T) {
	// Setup
	m := NewMaster()
	inputFiles := []string{"input1.txt", "input2.txt"}
	nReduce := 2
	m.CreateMapTasks(inputFiles, nReduce)

	// Simulate workers for Map phase
	for i := 1; i <= len(inputFiles); i++ {
		workerID := i

		// Worker requests a task
		reqArgs := &types.RequestTaskArgs{WorkerId: workerID}
		reply := &types.RequestTaskReply{}
		err := m.RequestTask(reqArgs, reply)
		if err != nil {
			t.Fatalf("RequestTask for map failed: %v", err)
		}
		if reply.Task == nil {
			t.Fatalf("Expected a map task but got nil")
		}
		if reply.Task.Type != types.MapTaskType {
			t.Errorf("Expected Map task, got %s", reply.Task.Type)
		}

		// Worker reports task completion
		reportArgs := &types.ReportTaskArgs{
			WorkerId: workerID,
			TaskId:   reply.Task.Id,
			TaskType: types.MapTaskType,
		}
		reportReply := &types.ReportTaskReply{}
		err = m.ReportTask(reportArgs, reportReply)
		if err != nil {
			t.Fatalf("ReportTask for map failed: %v", err)
		}
		if !reportReply.Ack {
			t.Fatalf("ReportTask for map was not acknowledged")
		}
	}

	// Check phase transition
	m.mu.Lock()
	if m.phase != ReducePhase {
		t.Errorf("Expected phase to be Reduce, but got %s", m.phase)
	}
	m.mu.Unlock()

	// Simulate workers for Reduce phase
	for i := 0; i < nReduce; i++ {
		workerID := i + 10 // Use different worker IDs for clarity

		// Worker requests a task
		reqArgs := &types.RequestTaskArgs{WorkerId: workerID}
		reply := &types.RequestTaskReply{}
		err := m.RequestTask(reqArgs, reply)
		if err != nil {
			t.Fatalf("RequestTask for reduce failed: %v", err)
		}
		if reply.Task == nil {
			t.Fatalf("Expected a reduce task but got nil")
		}
		if reply.Task.Type != types.ReduceTaskType {
			t.Errorf("Expected Reduce task, got %s", reply.Task.Type)
		}

		// Worker reports task completion
		reportArgs := &types.ReportTaskArgs{
			WorkerId: workerID,
			TaskId:   reply.Task.Id,
			TaskType: types.ReduceTaskType,
		}
		reportReply := &types.ReportTaskReply{}
		err = m.ReportTask(reportArgs, reportReply)
		if err != nil {
			t.Fatalf("ReportTask for reduce failed: %v", err)
		}
	}

	// Check for Done signal
	select {
	case <-m.Done():
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Master did not signal Done within 1 second")
	}

	m.mu.Lock()
	if m.phase != DonePhase {
		t.Errorf("Expected phase to be Done, but got %s", m.phase)
	}
	m.mu.Unlock()
}
