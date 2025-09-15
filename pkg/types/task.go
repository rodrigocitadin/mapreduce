package types

type KeyValue struct {
	Key   string
	Value string
}

type TaskType string

const (
	MapTaskType    TaskType = "MAP"
	ReduceTaskType TaskType = "REDUCE"
)

type TaskState string

const (
	IdleTaskState       TaskState = "IDLE"
	InProgressTaskState TaskState = "IN_PROGRESS"
	CompletedTaskState  TaskState = "COMPLETED"
)

type Task struct {
	Id       int
	WorkerId int
	NReduce  int    // R, the number of reduce tasks
	NMap     int    // M, the number of map tasks
	Filename string // Input file for map tasks
	Type     TaskType
	State    TaskState
}
