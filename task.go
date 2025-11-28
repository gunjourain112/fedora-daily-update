package main

type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusDone
	StatusError
)

type Task struct {
	Name    string
	Command string
	Args    []string
	Status  TaskStatus
	Output  string // To store the last few lines of output or full log
	Error   error
}
