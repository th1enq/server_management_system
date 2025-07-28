package configs

import (
	"fmt"
	"os"

	"github.com/th1enq/server_management_system/configs"
	"gopkg.in/yaml.v2"
)

type ConfigFilePath string

type Config struct {
	Server        Server        `yaml:"server"`
	Database      Database      `yaml:"database"`
	Cache         Cache         `yaml:"cache"`
	Log           Log           `yaml:"log"`
	JWT           JWT           `yaml:"jwt"`
	Cron          Cron          `yaml:"cron"`
	Elasticsearch ElasticSearch `yaml:"elasticsearch"`
	Email         Email         `yaml:"email"`
	TSDB          TSDB          `yaml:"tsdb"`
	MQ            MQ            `yaml:"mq"`
}

func NewConfig(filePath ConfigFilePath) (Config, error) {
	var (
		configBytes = configs.DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(string(filePath))
		if err != nil {
			return Config{}, fmt.Errorf("failed to read YAML file: %w", err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return config, nil
}

func Load() (Config, error) {
	return NewConfig("configs/config.dev.yaml")
}
