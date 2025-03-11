package config

import "fmt"

type Config struct {
	Validators   []Validator
	BlockWatcher BlockWatcherConfig `mapstructure:"block-watcher"`
	RPC          RPCConfig          `mapstructure:"rpc"`
	LogLevel     string             `mapstructure:"log-level"`
	HTTPServer   HTTPServerConfig   `mapstructure:"http-server"`
}

type Validator struct {
	Address  string `mapstructure:"address"`
	Name     string `mapstructure:"name"`
	Instance string `mapstructure:"instance"`
}

type BlockWatcherConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	RefreshInterval int  `mapstructure:"refresh-interval"`
}

type RPCConfig struct {
	Endpoint string `mapstructure:"endpoint"`
}

type HTTPServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (c *Config) Validate() error {
	if len(c.Validators) == 0 {
		return fmt.Errorf("no validators provided")
	}

	for _, v := range c.Validators {
		if v.Address == "" {
			return fmt.Errorf("validator address is required")
		}
		if v.Name == "" {
			return fmt.Errorf("validator name is required")
		}
		if v.Instance == "" {
			return fmt.Errorf("validator instance name is required")
		}
	}

	if c.BlockWatcher.RefreshInterval <= 0 {
		return fmt.Errorf("block watcher refresh interval must be greater than 0")
	}

	if c.RPC.Endpoint == "" {
		return fmt.Errorf("rpc endpoint is required")
	}

	return nil
}
