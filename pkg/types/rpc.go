package types

type RegisterWorkerArgs struct {
	Addr string
}
type RegisterWorkerReply struct {
	WorkerID int
}

type HeartbeatArgs struct {
	WorkerID int
}
type HeartbeatReply struct {
	Ack bool
}
