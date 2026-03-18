package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const ConfigPath = ".labctl/config.yml"
const ConfigEnvOverride = "LABCTL_CONFIG"

// Config contains the configuration for the labctl application.
type Config struct {
	APIToken          string        `yaml:"apiToken"`
	ProjectID         string        `yaml:"projectID"`
	SSHKeyFingerprint string        `yaml:"sshKeyFingerprint"`
	BaseImage         string        `yaml:"baseImage"`
	PollInterval      time.Duration `yaml:"pollInterval"`
	PollTimeout       time.Duration `yaml:"pollTimeout"`
	TagName           string        `yaml:"tagName"`
}

// NewConfig returns a new decoded Config struct.
func NewConfig(configPath string) (Config, error) {
	config := Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}
