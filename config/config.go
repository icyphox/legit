package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Git struct {
		ScanPath string   `yaml:"scanPath"`
		Readme   []string `yaml:"readme"`
	} `yaml:"git"`
	Template struct {
		Dir string `yaml:"dir"`
	} `yaml:"template"`
	Meta struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	} `yaml:"meta"`
	Server struct {
		Host string `yaml:"host"`
		Port int `yaml:"port"`
	} `yaml:"server"`
}

func Read(f string) (*Config, error) {
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	c := Config{}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &c, nil
}
