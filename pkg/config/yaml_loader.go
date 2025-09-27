package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type YAMLLoader struct{}

func (l *YAMLLoader) CanLoad(path string) bool {
	ext := GetConfigFileExtension(path)
	return ext == ".yaml" || ext == ".yml"
}

func (l *YAMLLoader) Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	data = []byte(expandEnvVars(string(data)))

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing YAML config file: %w", err)
	}

	return &config, nil
}

func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$(\w+)`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := strings.TrimPrefix(match, "${")
		varName = strings.TrimPrefix(varName, "$")
		varName = strings.TrimSuffix(varName, "}")

		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match
	})
}