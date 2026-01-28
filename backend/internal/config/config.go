package config

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Judge    JudgeConfig    `yaml:"judge"`
	AI       AIConfig       `yaml:"ai"`
	Paths    PathsConfig    `yaml:"paths"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}

type JudgeConfig struct {
	Sandbox string `yaml:"sandbox"`
	Workers int    `yaml:"workers"`
	Timeout int    `yaml:"timeout"`
}

type AIConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	APIURL   string `yaml:"api_url"`
	Model    string `yaml:"model"`
	Timeout  int    `yaml:"timeout"`
}

type PathsConfig struct {
	Problems    string `yaml:"problems"`
	Submissions string `yaml:"submissions"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	Expire time.Duration `yaml:"expire"`
}

var GlobalConfig *Config

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// 替换环境变量
	content := string(data)
	content = replaceEnvVars(content)

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Judge.Workers == 0 {
		cfg.Judge.Workers = 2
	}
	if cfg.Judge.Timeout == 0 {
		cfg.Judge.Timeout = 30
	}
	if cfg.JWT.Expire == 0 {
		cfg.JWT.Expire = 72 * time.Hour
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

// replaceEnvVars 替换配置中的环境变量 ${VAR_NAME}
func replaceEnvVars(content string) string {
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			placeholder := "${" + parts[0] + "}"
			content = strings.ReplaceAll(content, placeholder, parts[1])
		}
	}
	return content
}
