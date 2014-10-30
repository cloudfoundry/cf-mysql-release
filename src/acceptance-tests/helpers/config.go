package helpers

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

type Plan struct {
	Name               string `json:"plan_name"`
	MaxStorageMb       int    `json:"max_storage_mb"`
	MaxUserConnections int    `json:"max_user_connections"`
}

type MysqlIntegrationConfig struct {
	IntegrationConfig
	SmokeTestsOnly        bool   `json:"smoke_tests_only"`
	ExcludeDashboardTests bool   `json:"exclude_dashboard_tests"`
	BrokerHost            string `json:"broker_host"`
	ServiceName           string `json:"service_name"`
	Plans                 []Plan `json:"plans"`
}

func LoadConfig() (config MysqlIntegrationConfig) {
	path := os.Getenv("CONFIG")
	if path == "" {
		panic("Must set $CONFIG to point to an integration config .json file.")
	}

	return LoadPath(path)
}

func LoadPath(path string) (config MysqlIntegrationConfig) {
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
