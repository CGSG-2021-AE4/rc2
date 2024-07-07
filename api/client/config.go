package client

import (
	"encoding/json"
	"os"
)

type ScriptDescriptor struct {
	Name string `json:"name"`
	File string `json:"file"`
}

type Config struct {
	HomeDir  string             `json:"homeDir"`
	URL      string             `json:"url"`
	Login    string             `json:"login"`
	Password string             `json:"password"`
	Scripts  []ScriptDescriptor `json:"scripts"`
}

// Config
func LoadConfig(filename string) (*Config, error) {

	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
