package config

import (
	"fmt"
	"time"

	uberConfig "go.uber.org/config"
)

const (
	configFile = "base.yaml"
	configPath = "/root/"
)

type Logger struct {
	Level    string
	Encoding string
	Colored  bool
	Tags     []string
}

type TCP struct {
	Host             string        `yaml:"host"`
	Port             uint16        `yaml:"port"`
	ReqTimeout       time.Duration `yaml:"reqTimeout"`
	MaxReqSizeBytes  int           `yaml:"maxReqSizeBytes"`
	MaxReqPerSession int           `yaml:"maxReqPerSession"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
}

type Config struct {
	TCP    TCP    `yaml:"tcp"`
	Redis  Redis  `yaml:"redis"`
	Logger Logger `yaml:"logger"`
}

func New() (Config, error) {
	yml, err := uberConfig.NewYAML(
		uberConfig.File(fmt.Sprintf("%s%s", configPath, configFile)),
	)
	if err != nil {
		return Config{}, fmt.Errorf("create new yaml provider failed: %w", err)
	}

	var config Config
	if err = yml.Get("").Populate(&config); err != nil {
		return Config{}, fmt.Errorf("unmarshal yaml failed: %w", err)
	}

	return config, nil
}
