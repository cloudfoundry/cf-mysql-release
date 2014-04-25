package cf_mysql_service

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/vito/cmdtest"

	"../helpers"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/runner"
)

func TestServices(t *testing.T) {
	helpers.SetupEnvironment(helpers.NewContext(IntegrationConfig))
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "P-MySQL Acceptance Tests", []Reporter{reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))})
}

func AppUri(appname string) string {
	return "http://" + appname + "." + IntegrationConfig.AppsDomain
}

func Curling(args ...string) func() *cmdtest.Session {
	return func() *cmdtest.Session {
		return Curl(args...)
	}
}

var IntegrationConfig = helpers.LoadConfig()
var sinatraPath = "../assets/sinatra_app"
