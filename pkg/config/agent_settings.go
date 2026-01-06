package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AgentSettings struct {
	ApiKey   string `json:"apiKey"`
	Security struct {
		Auth struct {
			SelectedType string `json:"selectedType"`
		} `json:"auth"`
	} `json:"security"`
	Tools struct {
		Sandbox interface{} `json:"sandbox"`
	} `json:"tools"`
}

func LoadAgentSettings(path string) (*AgentSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var settings AgentSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

func SaveAgentSettings(path string, settings *AgentSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func GetAgentSettings() (*AgentSettings, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".gemini", "settings.json")
	return LoadAgentSettings(path)
}
