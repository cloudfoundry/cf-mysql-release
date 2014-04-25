package helpers

import (
	"fmt"
	"time"

	ginkgoconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/cf-test-helpers/cf"
	"github.com/vito/cmdtest"
	. "github.com/vito/cmdtest/matchers"
)

type ConfiguredContext struct {
	config IntegrationConfig

	organizationName string
	spaceName        string

	regularUserUsername string
	regularUserPassword string

	isPersistent bool

	existingServiceInstanceOrgName string
	existingServiceInstanceSpaceName string
}

func NewContext(config IntegrationConfig) *ConfiguredContext {
	node := ginkgoconfig.GinkgoConfig.ParallelNode
	timeTag := time.Now().Format("2006_01_02-15h04m05.999s")

	return &ConfiguredContext{
		config: config,

		organizationName: fmt.Sprintf("MySqlATS-ORG-%d-%s", node, timeTag),
		spaceName:        fmt.Sprintf("MySQLATS-SPACE-%d-%s", node, timeTag),

		regularUserUsername: fmt.Sprintf("MySqlATS-USER-%d-%s", node, timeTag),
		regularUserPassword: "meow",

		isPersistent: false,

		existingServiceInstanceOrgName: config.ExistingServiceInstanceOrg,
		existingServiceInstanceSpaceName: config.ExistingServiceInstanceSpace,
	}
}

func (context *ConfiguredContext) Setup() {
	cf.AsUser(context.AdminUserContext(), func() {
			Expect(cf.Cf("create-user", context.regularUserUsername, context.regularUserPassword)).To(SayBranches(
				cmdtest.ExpectBranch{"OK", func() {}},
				cmdtest.ExpectBranch{"scim_resource_already_exists", func() {}},
			))

			Expect(cf.Cf("create-org", context.organizationName)).To(ExitWithTimeout(0, 60*time.Second))
		})
}

func (context *ConfiguredContext) Teardown() {
	cf.AsUser(context.AdminUserContext(), func() {
			Expect(cf.Cf("delete-user", "-f", context.regularUserUsername)).To(ExitWithTimeout(0, 60*time.Second))

			if !context.isPersistent {
				Expect(cf.Cf("delete-org", "-f", context.organizationName)).To(ExitWithTimeout(0, 60*time.Second))
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

func (context *ConfiguredContext) RegularUserExistingOrgSpaceContext() cf.UserContext {
	return cf.NewUserContext(
		context.config.ApiEndpoint,
		context.regularUserUsername,
		context.regularUserPassword,
		context.existingServiceInstanceOrgName,
		context.existingServiceInstanceSpaceName,
		context.config.SkipSSLValidation,
	)
}
