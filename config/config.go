package config

import (
	"errors"
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Token     string `yaml:"Token"`
	Access    string `yaml:"Access"`
	DriveType string `yaml:"DriveType"`
	DriveID   string `yaml:"DriveID"`
}

var configPath = "/etc/aliyun/aliyun.yaml"
var c *Config

func C() *Config {
	return c
}

func LoadYaml() {
	data, err := os.ReadFile(configPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		c = new(Config)
		SaveYaml()
		return
	}
	c = new(Config)
	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Fatalf("无法解析 YAML 数据: %v", err)
	}
}

func SaveYaml() error {
	updatedData, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, updatedData, 0644)
}
