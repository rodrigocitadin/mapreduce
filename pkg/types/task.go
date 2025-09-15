package types

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
	NReduce  int
	Filename string
	Type     TaskType
	State    TaskState
}
