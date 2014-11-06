package helpers

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	context_setup "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

func PrepareAndRunTests(packageName string, integrationConfig *MysqlIntegrationConfig, t *testing.T) {
	if integrationConfig.SmokeTestsOnly {
		ginkgoconfig.GinkgoConfig.FocusString = "Service instance lifecycle"
	}

	var skipStrings []string

	if ginkgoconfig.GinkgoConfig.SkipString != "" {
		skipStrings = append(skipStrings, ginkgoconfig.GinkgoConfig.SkipString)
	}

	if !integrationConfig.IncludeDashboardTests {
		skipStrings = append(skipStrings, "CF Mysql Dashboard")
	}

	if !integrationConfig.IncludeFailoverTests {
		skipStrings = append(skipStrings, "CF MySQL Failover")
	}

	if len(skipStrings) > 0 {
		ginkgoconfig.GinkgoConfig.SkipString = strings.Join(skipStrings, "|")
	}

	context_setup.TimeoutScale = integrationConfig.TimeoutScale
	context_setup.SetupEnvironment(context_setup.NewContext(integrationConfig.IntegrationConfig, "MySQLATS"))

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("P-MySQL Acceptance Tests -- %s", packageName), []Reporter{junitReporter})
}
