package config

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	httpkit "github.com/tiamxu/kit/http"
	"github.com/tiamxu/kit/log"
	"gopkg.in/yaml.v3"
)

var (
	cfg        *Config
	name       = "ai-agent"
	configPath = "config/config.yaml"
)

type Config struct {
	ENV      string                  `yaml:"env"`
	LogLevel string                  `yaml:"log_level"`
	HttpSrv  httpkit.GinServerConfig `yaml:"http_srv"`
	Aliyun   AliyunConfig            `yaml:"aliyun"`
	Doubao   DoubaoConfig            `yaml:"doubao"`
}
type AliyunConfig struct {
	BaseURL        string `yaml:"base_url"`
	APIKey         string `yaml:"api_key"`
	LLMModel       string `yaml:"llm_model"`
	EmbeddingModel string `yaml:"embedding_model"`
}

type DoubaoConfig struct {
	BaseURL        string `yaml:"base_url"`
	APIKey         string `yaml:"api_key"`
	LLMModel       string `yaml:"llm_model"`
	EmbeddingModel string `yaml:"embedding_model"`
}

func (c *Config) Initial() (err error) {

	defer func() {
		if err == nil {
			log.Printf("config initialed, env: %s,name: %s", cfg.ENV, name)
		}
	}()

	if level, err := logrus.ParseLevel(c.LogLevel); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	} else {
		log.DefaultLogger().SetLevel(level)
	}
	return nil
}
func LoadConfig() (*Config, error) {
	cfg = new(Config)
	env := "local"

	switch env {
	case "dev":
		configPath = "config/config-dev.yaml"
	case "test":
		configPath = "config/config-test.yaml"
	case "prod":
		configPath = "config/config-prod.yaml"
	default:
		configPath = "config/config.yaml"
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败:%w", err)
	}

	if err := yaml.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return cfg, nil

}
