package helpers

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
)

type ConfiguredContext struct {
	config IntegrationConfig

	organizationName string
	spaceName        string

	quotaDefinitionName string
	quotaDefinitionGUID string

	regularUserUsername string
	regularUserPassword string

	isPersistent bool
}

type quotaDefinition struct {
	Name string `json:"name"`

	NonBasicServicesAllowed bool `json:"non_basic_services_allowed"`

	TotalServices int `json:"total_services"`
	TotalRoutes   int `json:"total_routes"`

	MemoryLimit int `json:"memory_limit"`
}

func NewContext(config IntegrationConfig) *ConfiguredContext {
	node := ginkgoconfig.GinkgoConfig.ParallelNode
	timeTag := time.Now().Format("2006_01_02-15h04m05.999s")

	return &ConfiguredContext{
		config: config,

		quotaDefinitionName: fmt.Sprintf("MySqlATS-QUOTA-%d-%s", node, timeTag),

		organizationName: fmt.Sprintf("MySqlATS-ORG-%d-%s", node, timeTag),
		spaceName:        fmt.Sprintf("MySQLATS-SPACE-%d-%s", node, timeTag),

		regularUserUsername: fmt.Sprintf("MySqlATS-USER-%d-%s", node, timeTag),
		regularUserPassword: "meow",

		isPersistent: false,
	}
}

func (context *ConfiguredContext) Setup() {
	cf.AsUser(context.AdminUserContext(), func() {
		channel := cf.Cf("create-user", context.regularUserUsername, context.regularUserPassword)
		select {
		case <-channel.Out.Detect("OK"):
		case <-channel.Out.Detect("scim_resource_already_exists"):
		case <-time.After(10 * time.Second):
			Fail("failed to create user")
		}

		definition := quotaDefinition{
			Name: context.quotaDefinitionName,

			TotalServices: 100,
			TotalRoutes:   1000,

			MemoryLimit: 10240,

			NonBasicServicesAllowed: true,
		}

		definitionPayload, err := json.Marshal(definition)
		Expect(err).ToNot(HaveOccurred())

		var response cf.GenericResource

		cf.ApiRequest("POST", "/v2/quota_definitions", &response, string(definitionPayload))

		context.quotaDefinitionGUID = response.Metadata.Guid

		Eventually(cf.Cf("create-org", context.organizationName), 60*time.Second).Should(Exit(0))
		Eventually(cf.Cf("set-quota", context.organizationName, definition.Name), 60*time.Second).Should(Exit(0))
	})
}

func (context *ConfiguredContext) Teardown() {
	cf.AsUser(context.AdminUserContext(), func() {
		Eventually(cf.Cf("delete-user", "-f", context.regularUserUsername), 60*time.Second).Should(Exit(0))

		if !context.isPersistent {
			Eventually(cf.Cf("delete-org", "-f", context.organizationName), 60*time.Second).Should(Exit(0))

			cf.ApiRequest(
				"DELETE",
				"/v2/quota_definitions/"+context.quotaDefinitionGUID+"?recursive=true",
				nil,
			)
		}
	})
}

func (context *ConfiguredContext) AdminUserContext() cf.UserContext {
	return cf.NewUserContext(
		context.config.ApiEndpoint,
		context.config.AdminUser,
		context.config.AdminPassword,
		"",
		"",
		context.config.SkipSSLValidation,
	)
}

func (context *ConfiguredContext) RegularUserContext() cf.UserContext {
	return cf.NewUserContext(
		context.config.ApiEndpoint,
		context.regularUserUsername,
		context.regularUserPassword,
		context.organizationName,
		context.spaceName,
		context.config.SkipSSLValidation,
	)
}
