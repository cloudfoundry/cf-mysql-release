package service_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"../../helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var IntegrationConfig = helpers.LoadConfig()

func TestService(t *testing.T) {
	helpers.PrepareAndRunTests("Service", &IntegrationConfig, t)
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)
})

func AppUri(appname string) string {
	return "http://" + appname + "." + IntegrationConfig.AppsDomain
}

func Curling(args ...string) func() *gexec.Session {
	return func() *gexec.Session {
		return Curl(args...)
	}
}
