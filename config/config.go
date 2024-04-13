package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Redis  Redis  `yaml:"redis"`
	Serial Serial `yaml:"serial"`
}

type Redis struct {
	DSN string `yaml:"dsn"`
}

type Serial struct {
	PortName string `yaml:"portName"`
}

func New() (config Config, err error) {
	fileName, err := filepath.Abs("./config.yaml")
	if err != nil {
		return config, err
	}

	err = loadYamlFile(&config, fileName)
	if err != nil {
		return config, err
	}

	return config, nil
}

func loadYamlFile(cfg *Config, fileName string) error {
	_, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	fs, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer fs.Close()

	err = yaml.NewDecoder(fs).Decode(cfg)
	if err != nil {
		return err
	}

	return nil
}
