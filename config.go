package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const stateFile = "/opt/vessel/apps.yaml"

type AppState struct {
	Apps []*AppConfig `yaml:"apps"`
}

func parseYAML(data []byte, cfg *AppConfig) error {
	return yaml.Unmarshal(data, cfg)
}

func SaveApp(cfg *AppConfig) error {
	state, _ := loadState()

	for i, a := range state.Apps {
		if a.Name == cfg.Name {
			state.Apps[i] = cfg
			return writeState(state)
		}
	}
	state.Apps = append(state.Apps, cfg)
	return writeState(state)
}

func ListApps() ([]*AppConfig, error) {
	state, err := loadState()
	if err != nil {
		return nil, err
	}
	return state.Apps, nil
}

func RemoveApp(name string) error {
	state, err := loadState()
	if err != nil {
		return err
	}

	if err := StopService(name); err != nil {
		fmt.Printf("warning: could not stop service: %v\n", err)
	}

	// Remove unit file.
	unitPath := filepath.Join(systemdDir, fmt.Sprintf("vessel-%s.service", name))
	os.Remove(unitPath)

	// Remove proxy config.
	os.Remove(filepath.Join(caddyConfigDir, fmt.Sprintf("vessel-%s.caddy", name)))
	os.Remove(filepath.Join("/etc/nginx/sites-enabled", fmt.Sprintf("vessel-%s", name)))
	os.Remove(filepath.Join("/etc/nginx/sites-available", fmt.Sprintf("vessel-%s", name)))

	// Remove app directory.
	os.RemoveAll(filepath.Join(deployRoot, "apps", name))

	// Remove from state.
	var filtered []*AppConfig
	for _, a := range state.Apps {
		if a.Name != name {
			filtered = append(filtered, a)
		}
	}
	state.Apps = filtered

	fmt.Printf("✅ Removed %s\n", name)
	return writeState(state)
}

func loadState() (*AppState, error) {
	data, err := os.ReadFile(stateFile)
	if os.IsNotExist(err) {
		return &AppState{}, nil
	}
	if err != nil {
		return nil, err
	}
	var s AppState
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func writeState(state *AppState) error {
	os.MkdirAll(filepath.Dir(stateFile), 0755)
	data, err := yaml.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(stateFile, data, 0644)
}