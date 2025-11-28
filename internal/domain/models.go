package domain

// TaskType은 태스크의 종류(내장/사용자정의)를 구분합니다.
type TaskType int

const (
	TaskTypeBuiltin TaskType = iota
	TaskTypeCustom
)

// Task는 실행 가능한 업데이트 작업 단위입니다.
type Task struct {
	ID      string   // 고유 식별자 (예: "dnf", "flatpak", "custom-1")
	Name    string   // 화면에 표시될 이름 (예: "시스템 업데이트 (dnf)")
	Command string   // 실행할 명령어
	Args    []string // 명령어 인자
	Type    TaskType // 태스크 타입
}

// Config는 애플리케이션 설정을 나타냅니다.
type Config struct {
	CustomTasks []Task `json:"custom_tasks"`
}
