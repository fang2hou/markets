package config

import (
	"errors"

	yaml "gopkg.in/yaml.v3"
)

type configData struct {
	Exchanges  map[string]map[string]string `yaml:"exchange"`
	Currencies []string                     `yaml:"currency"`
	Log        struct {
		Enabled bool   `yaml:"enable"`
		Path    string `yaml:"path"`
	} `yaml:"log"`
}

// Config is the main configuration struct for arbitrary services.
type Config struct {
	loaded bool
	data   configData
}

func (c *Config) Load(dataBytes []byte) error {
	if err := yaml.Unmarshal(dataBytes, &c.data); err != nil {
		return err
	}

	c.loaded = true
	return nil
}

func (c *Config) GetExchangeSetting(exchangeName string) (map[string]string, error) {
	if !c.loaded {
		return nil, errors.New("config has not been loaded")
	}

	if c.data.Exchanges == nil {
		return nil, errors.New("no exchanges found in config")
	}

	if exchangeSetting, ok := c.data.Exchanges[exchangeName]; ok {
		return exchangeSetting, nil
	} else {
		return nil, errors.New("No exchange found with name " + exchangeName)
	}
}

func (c *Config) GetCurrenciesSetting() ([]string, error) {
	if !c.loaded {
		return nil, errors.New("config has not been loaded")
	}

	if c.data.Currencies == nil {
		return nil, errors.New("no currencies found in config")
	}

	return c.data.Currencies, nil
}
