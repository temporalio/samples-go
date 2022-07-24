package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/validator.v2"
	"gopkg.in/yaml.v3"
)

const (
	EnvKeyConfigHome  = "TEMPORAL_CONFIG_ROOT"
	EnvKeyEnvironment = "TEMPORAL_ENVIRONMENT"
	EnvKeyConfigFile  = "TEMPORAL_CONFIG_FILE"

	EnvDefaultConfigHome  = ".temporal"
	EnvDefaultEnvironment = "local"
	EnvDefaultConfigFile  = "config.yaml"
)

func loadConfig(configRoot string, config interface{}) error {
	configFile := getEnvOrDefaultString(EnvKeyConfigFile, EnvDefaultConfigFile)

	configPath := filepath.Join(configRoot, configFile)
	fmt.Printf("Loading config from: %v\n", configPath)

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	return validator.Validate(config)
}

func getEnvOrDefaultString(envVarName string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(envVarName))
	if value == "" {
		value = defaultValue
	}

	return value
}
