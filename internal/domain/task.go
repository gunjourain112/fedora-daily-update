package domain

// TaskStatus represents the current state of a task.
type TaskStatus int

const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusDone
	StatusError
)

// Task represents a unit of work to be executed.
type Task struct {
	Name    string
	Command string
	Args    []string
	Status  TaskStatus
	Output  string
	Error   error
	IsSystem bool // True if it's a built-in system task (read-only in settings)
}

// TaskManager aggregates system and custom tasks.
type TaskManager struct {
	// dependencies could go here (e.g. config manager)
}

// NewTaskManager creates a new task manager.
func NewTaskManager() *TaskManager {
	return &TaskManager{}
}

// GetTasks returns the full list of tasks to be executed.
// It combines system tasks (dnf, flatpak) and custom tasks provided as arguments.
// In a real scenario, this might pull directly from config, but passing data is cleaner.
func (tm *TaskManager) GetTasks(customTasks []Task) []Task {
	// System Tasks (Fixed order, usually first or last depending on preference.
	// User mentioned "dnf and flatpak are defaults". Let's put them first.)

	// However, the user also mentioned "Custom Menu Management" only for custom tasks.
	// So we need to distinguish them.

	tasks := []Task{
		{
			Name:     "시스템 업데이트 (dnf)",
			Command:  "dnf",
			Args:     []string{"update", "-y"},
			IsSystem: true,
		},
		{
			Name:     "Flatpak 업데이트",
			Command:  "flatpak",
			Args:     []string{"update", "-y"},
			IsSystem: true,
		},
	}

	tasks = append(tasks, customTasks...)
	return tasks
}
