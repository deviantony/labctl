package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const CONFIG_PATH = ".labctl/config.yml"
const CONFIG_ENV_OVERRIDE = "LABCTL_CONFIG"

type Config struct {
	LXD LXDConfig          `yaml:"lxd"`
	DO  DigitalOceanConfig `yaml:"do"`
}

type LXDConfig struct {
	Server struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
	} `yaml:"server"`

	Key          string `yaml:"key"`
	Cert         string `yaml:"cert"`
	SSHPublicKey string `yaml:"sshPublicKey"`
}

type DigitalOceanConfig struct {
	APIToken          string        `yaml:"apiToken"`
	ProjectID         string        `yaml:"projectID"`
	SSHKeyFingerprint string        `yaml:"sshKeyFingerprint"`
	BaseImage         string        `yaml:"baseImage"`
	PollInterval      time.Duration `yaml:"pollInterval"`
	PollTimeout       time.Duration `yaml:"pollTimeout"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (Config, error) {
	// Create config structure
	config := Config{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	// Init new YAML decode
	d := yaml.NewDecoder(file)

	// Start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}
