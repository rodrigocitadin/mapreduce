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
