package helpers

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/cf-test-helpers/cf"
)

type ConfiguredContext struct {
	config IntegrationConfig

	organizationName string
	spaceName        string

	regularUserUsername string
	regularUserPassword string

	isPersistent bool
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

		Eventually(cf.Cf("create-org", context.organizationName), 60*time.Second).Should(Exit(0))
	})
}

func (context *ConfiguredContext) Teardown() {
	cf.AsUser(context.AdminUserContext(), func() {
		Eventually(cf.Cf("delete-user", "-f", context.regularUserUsername), 60*time.Second).Should(Exit(0))

		if !context.isPersistent {
			Eventually(cf.Cf("delete-org", "-f", context.organizationName), 60*time.Second).Should(Exit(0))
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
