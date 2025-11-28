package config

import (
	"os"
	"path/filepath"
	"testing"

	"my-updater/internal/domain"
)

func TestLoadConfig_NoFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	manager := &Manager{configPath: filepath.Join(tmpDir, "config.json")}
	cfg, err := manager.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.CustomTasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(cfg.CustomTasks))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	manager := &Manager{configPath: filepath.Join(tmpDir, "config.json")}

	tasks := []domain.Task{
		{ID: "t1", Name: "Test Task", Command: "echo", Args: []string{"hello"}, Type: domain.TaskTypeCustom},
	}
	cfg := &domain.Config{CustomTasks: tasks}

	if err := manager.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loadedCfg, err := manager.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(loadedCfg.CustomTasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedCfg.CustomTasks))
	}
	if loadedCfg.CustomTasks[0].Name != "Test Task" {
		t.Errorf("Expected 'Test Task', got '%s'", loadedCfg.CustomTasks[0].Name)
	}
}
