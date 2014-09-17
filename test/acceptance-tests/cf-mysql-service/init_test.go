package cf_mysql_service

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"../helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	context_setup "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

func TestServices(t *testing.T) {
	if IntegrationConfig.SmokeTestsOnly {
		ginkgoconfig.GinkgoConfig.FocusString = "Service instance lifecycle"
	}

	context_setup.TimeoutScale = IntegrationConfig.TimeoutScale

	context_setup.SetupEnvironment(context_setup.NewContext(IntegrationConfig.IntegrationConfig, "MySQLATS"))
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "P-MySQL Acceptance Tests", []Reporter{reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))})
}

func AppUri(appname string) string {
	return "http://" + appname + "." + IntegrationConfig.AppsDomain
}

func Curling(args ...string) func() *gexec.Session {
	return func() *gexec.Session {
		return Curl(args...)
	}
}

var IntegrationConfig = helpers.LoadConfig()
var sinatraPath = "../assets/sinatra_app"
