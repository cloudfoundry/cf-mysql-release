package service_config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	ConfigPathEnvVar = "CONFIG_PATH"
	ConfigEnvVar     = "CONFIG"
)

var NoConfigError = errors.New("No Config or Config Path Specified. Please supply one of the following: -config, -configPath, CONFIG, or CONFIG_PATH")

type ServiceConfig struct {
	configFlag     string
	configPathFlag string
	flagSet        *flag.FlagSet
	defaultModel   interface{}
	helpWriter     io.Writer
}

func New() *ServiceConfig {
	return &ServiceConfig{
		helpWriter: os.Stderr,
	}
}

func (c *ServiceConfig) AddFlags(flagSet *flag.FlagSet) {
	c.flagSet = flagSet
	c.flagSet.StringVar(&c.configFlag, "config", "", "json encoded configuration string")
	c.flagSet.StringVar(&c.configPathFlag, "configPath", "", "path to configuration file with json encoded content")

	c.flagSet.SetOutput(c.helpWriter)
	c.flagSet.Usage = func() {
		c.PrintUsage()
	}
}

func (c ServiceConfig) ConfigBytes() ([]byte, error) {
	if c.configFlag != "" {
		return []byte(c.configFlag), nil
	}

	if c.configPathFlag != "" {
		absolutePath, err := filepath.Abs(c.configPathFlag)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Making config file path absolute: %s", err.Error()))
		}

		bytes, err := ioutil.ReadFile(absolutePath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Reading config file: %s", err.Error()))
		}

		return bytes, nil
	}

	config := os.Getenv(ConfigEnvVar)
	if config != "" {
		return []byte(config), nil
	}

	configPath := os.Getenv(ConfigPathEnvVar)
	if configPath != "" {
		absolutePath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Making config file path absolute: %s", err.Error()))
		}

		bytes, err := ioutil.ReadFile(absolutePath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Reading config file: %s", err.Error()))
		}

		return bytes, nil
	}

	return nil, NoConfigError
}

func (c ServiceConfig) ConfigPath() string {
	return c.configPathFlag
}

func (c ServiceConfig) Read(model interface{}) error {
	bytes, err := c.ConfigBytes()
	if err != nil {
		return err
	}

	reader := NewReader(bytes)

	if c.defaultModel != nil {
		err = reader.ReadWithDefaults(model, c.defaultModel)
	} else {
		err = reader.Read(model)
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceConfig) AddDefaults(defaultModel interface{}) {
	c.defaultModel = defaultModel
}

func (c ServiceConfig) PrintUsage() {
	fmt.Fprint(c.helpWriter, "Expected usage:\n")
	c.flagSet.PrintDefaults()

	if c.defaultModel != nil {
		defaultStr, err := yaml.Marshal(c.defaultModel)
		if err != nil {
			fmt.Fprintf(c.helpWriter, "Error printing defaults: %v", err)
		} else {
			fmt.Fprintf(c.helpWriter, "Default config values:\n%s", defaultStr)
		}
	}
}
