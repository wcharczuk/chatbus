package server

import (
	"sync"

	"github.com/wcharczuk/go-web/config"
)

var (
	_configLock sync.Mutex
	_config     *AppConfig
)

// DefaultConfig returns the default / cached config.
func DefaultConfig() *AppConfig {
	if _config == nil {
		_configLock.Lock()
		defer _configLock.Unlock()
		if _config == nil {
			_config = &AppConfig{}
		}
	}
	return _config
}

// SetDefaultConfig lets you set the default config.
func SetDefaultConfig(config *AppConfig) {
	_configLock.Lock()
	defer _configLock.Unlock()
	_config = config
}

// AppConfig is the configuration for the app.
type AppConfig struct {
	AppName     string `env:"APP_NAME" env_default:"Chat Bus"`
	Environment string `env:"ENV" env_default:"dev"`
	Port        string `env:"PORT" env_default:"8080"`
}

// FromEnvironment reads the config from the environment.
func (c *AppConfig) FromEnvironment() error {
	return config.FromEnvironment(c)
}
