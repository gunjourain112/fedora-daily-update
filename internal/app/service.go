package app

import (
	"my-updater/internal/config"
	"my-updater/internal/domain"
)

// TaskService는 태스크 목록을 제공하고 관리하는 비즈니스 로직을 담당합니다.
type TaskService struct {
	configManager *config.Manager
}

// NewTaskService는 TaskService를 생성합니다.
func NewTaskService(cm *config.Manager) *TaskService {
	return &TaskService{
		configManager: cm,
	}
}

// GetAllTasks는 기본 태스크와 사용자 정의 태스크를 합쳐서 반환합니다.
func (s *TaskService) GetAllTasks() ([]domain.Task, error) {
	tasks := []domain.Task{
		{
			ID:      "dnf",
			Name:    "시스템 패키지 업데이트 (dnf)",
			Command: "sudo",
			Args:    []string{"dnf", "update", "-y"},
			Type:    domain.TaskTypeBuiltin,
		},
		{
			ID:      "flatpak",
			Name:    "Flatpak 애플리케이션 업데이트",
			Command: "flatpak",
			Args:    []string{"update"},
			Type:    domain.TaskTypeBuiltin,
		},
	}

	cfg, err := s.configManager.LoadConfig()
	if err != nil {
		return nil, err
	}

	for i, ct := range cfg.CustomTasks {
		// ID가 없으면 임의로 생성 (혹은 저장 시 생성)
		if ct.ID == "" {
			ct.ID = "custom-" + ct.Name // 간단한 ID 생성 예시
		}
		ct.Type = domain.TaskTypeCustom
		tasks = append(tasks, ct)
		// 실제로는 cfg의 데이터를 그대로 쓰되 Type을 명시
		_ = i
	}

	return tasks, nil
}

// SaveCustomTasks는 사용자 정의 태스크 목록을 저장합니다.
func (s *TaskService) SaveCustomTasks(tasks []domain.Task) error {
	// Custom 타입만 필터링하여 저장
	customTasks := []domain.Task{}
	for _, t := range tasks {
		if t.Type == domain.TaskTypeCustom {
			customTasks = append(customTasks, t)
		}
	}

	cfg := &domain.Config{
		CustomTasks: customTasks,
	}

	return s.configManager.SaveConfig(cfg)
}
