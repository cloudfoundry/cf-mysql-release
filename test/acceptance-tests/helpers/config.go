package helpers

import (
	"encoding/json"
	"fmt"
	"os"
)

type Plan struct {
	Name         string `json:"plan_name"`
	MaxStorageMb int    `json:"max_storage_mb"`
    MaxUserConnections		int `json:"max_user_connections"`
}

type IntegrationConfig struct {
	AppsDomain         string  `json:"apps_domain"`
	ApiEndpoint        string  `json:"api_url"`
	AdminUser          string  `json:"admin_user"`
	AdminPassword      string  `json:"admin_password"`
	SkipSSLValidation  bool    `json:"skip_ssl_validation"`
	BrokerHost         string  `json:"broker_host"`
	ServiceName        string  `json:"service_name"`
	Plans              []Plan  `json:"plans"`
	MaxUserConnections int     `json:"max_user_connections"`
	SmokeTestsOnly     bool    `json:"smoke_tests_only"`
	TimeoutScale       float64 `json:"timeout_scale"`
}

func LoadConfig() (config IntegrationConfig) {
	path := os.Getenv("CONFIG")
	if path == "" {
		panic("Must set $CONFIG to point to an integration config .json file.")
	}

	return LoadPath(path)
}

func LoadPath(path string) (config IntegrationConfig) {
	configFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	if config.ApiEndpoint == "" {
		panic("missing configuration 'api'")
	}

	if config.AdminUser == "" {
		panic("missing configuration 'admin_user'")
	}

	if config.ApiEndpoint == "" {
		panic("missing configuration 'admin_password'")
	}

	if config.ServiceName == "" {
		panic("missing configuration 'service_name'")
	}

	if config.Plans == nil {
		panic("missing configuration 'plans'")
	}

	for index, plan := range config.Plans {
		if plan.Name == "" {
			panic(fmt.Sprintf("missing configuration 'plans.name' for plan %d", index))
		}

		if plan.MaxStorageMb == 0 {
			panic(fmt.Sprintf("missing configuration 'plans.max_storage_mb' for plan %d", index))
		}

		if plan.MaxUserConnections == 0 {
			panic(fmt.Sprintf("missing configuration 'plans.max_user_connections' for plan %d", index))
		}
	}

	if config.BrokerHost == "" {
		panic("missing configuration 'broker_host'")
	}

	if config.TimeoutScale <= 0 {
		config.TimeoutScale = 1
	}
	return
}
