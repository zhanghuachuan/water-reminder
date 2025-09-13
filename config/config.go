package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type OperatorConfig struct {
	Name      string                 `yaml:"name"`
	Type      string                 `yaml:"type"`
	Config    map[string]interface{} `yaml:"config"`
	DependsOn []string               `yaml:"depends_on"`
}

type AppConfig struct {
	Operators []OperatorConfig `yaml:"operators"`
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
