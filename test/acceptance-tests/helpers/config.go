package helpers

import (
	"encoding/json"
	"os"
)

type IntegrationConfig struct {
	AppsDomain        string `json:"apps_domain"`
	ApiEndpoint       string `json:"api_url"`
	AdminUser         string `json:"admin_user"`
	AdminPassword     string `json:"admin_password"`
	SkipSSLValidation bool   `json:"skip_ssl_validation"`
	BrokerHost 				string `json:"broker_host"`
	ServiceName				string `json:"service_name"`
	PlanName					string `json:"plan_name"`
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

	if config.PlanName == "" {
		panic("missing configuration 'plan_name'")
	}

	if config.BrokerHost == "" {
		panic("missing configuration 'broker_host'")
	}

	return
}
