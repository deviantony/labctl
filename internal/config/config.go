package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const CONFIG_PATH = ".labctl/config.yml"
const CONFIG_ENV_OVERRIDE = "LABCTL_CONFIG"

// Provider is an enum for the supported providers
type Provider string

const (
	// PROVIDER_DO is the DigitalOcean provider
	PROVIDER_DO Provider = "DigitalOcean"
	// PROVIDER_LXD is the LXD provider
	PROVIDER_LXD Provider = "LXD"
)

// Config contains the configuration for the labctl application
type Config struct {
	DO        DigitalOceanConfig `yaml:"do"`
	LXD       LXDConfig          `yaml:"lxd"`
	DockerHub DockerHubConfig    `yaml:"dockerhub"`

	provider Provider
}

// DockerHubConfig contains the DockerHub configuration
type DockerHubConfig struct {
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Timeout  time.Duration `yaml:"timeout"`
}

// DigitalOceanConfig contains the configuration for the DigitalOcean provider
type DigitalOceanConfig struct {
	APIToken          string        `yaml:"apiToken"`
	ProjectID         string        `yaml:"projectID"`
	SSHKeyFingerprint string        `yaml:"sshKeyFingerprint"`
	BaseImage         string        `yaml:"baseImage"`
	PollInterval      time.Duration `yaml:"pollInterval"`
	PollTimeout       time.Duration `yaml:"pollTimeout"`
	TagName           string        `yaml:"tagName"`
}

// LXDConfig contains the configuration for the LXD provider
type LXDConfig struct {
	Server struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
	} `yaml:"server"`

	Client struct {
		Cert    string        `yaml:"cert"`
		Key     string        `yaml:"key"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"client"`

	SSHPublicKey string `yaml:"sshPublicKey"`
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

// GetProvider returns the provider
func (c *Config) GetProvider() Provider {
	return c.provider
}

// SetProvider sets the provider
func (c *Config) SetProvider(provider Provider) {
	c.provider = provider
}
