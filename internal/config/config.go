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

type Tcp struct {
	Host            string        `yaml:"host"`
	Port            uint16        `yaml:"port"`
	ReqTimeout      time.Duration `yaml:"reqTimeout"`
	MaxReqSizeBytes int           `yaml:"maxReqSizeBytes"`
}

type Redis struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
}

type Config struct {
	Tcp    Tcp    `yaml:"tcp"`
	Redis  Redis  `yaml:"redis"`
	Logger Logger `yaml:"logger"`
}

func New() (Config, error) {
	y, err := uberConfig.NewYAML(
		uberConfig.File(fmt.Sprintf("%s%s", configPath, configFile)),
	)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = y.Get("").Populate(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
