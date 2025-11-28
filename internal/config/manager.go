package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"my-updater/internal/domain"
)

// Manager는 설정 파일 입출력을 담당합니다.
type Manager struct {
	configPath string
}

// NewManager는 설정 매니저를 초기화합니다.
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "my-updater")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		configPath: filepath.Join(configDir, "config.json"),
	}, nil
}

// LoadConfig는 설정 파일을 읽어 반환합니다. 파일이 없으면 빈 설정을 반환합니다.
func (m *Manager) LoadConfig() (*domain.Config, error) {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return &domain.Config{CustomTasks: []domain.Task{}}, nil
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, err
	}

	var cfg domain.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// 파싱 에러 시 빈 설정 반환 (혹은 에러 처리)
		return &domain.Config{CustomTasks: []domain.Task{}}, nil
	}

	return &cfg, nil
}

// SaveConfig는 설정을 파일에 저장합니다.
func (m *Manager) SaveConfig(cfg *domain.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}
