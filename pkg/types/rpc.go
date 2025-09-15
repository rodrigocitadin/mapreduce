package types

type RegisterWorkerArgs struct {
	Addr string
}
type RegisterWorkerReply struct {
	WorkerId int
}

type HeartbeatArgs struct {
	WorkerId int
}
type HeartbeatReply struct {
	Ack bool
}

type RequestTaskArgs struct {
	WorkerId int
}

type RequestTaskReply struct {
	Task *Task
}

type ReportTaskArgs struct {
	WorkerId int
	TaskId   int
	TaskType TaskType
}

type ReportTaskReply struct {
	Ack bool
}
