package dashboard_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/dsl"

	"../../helpers"
)

var IntegrationConfig = helpers.LoadConfig()

func TestDashboard(t *testing.T) {
	helpers.PrepareAndRunTests("Dashboard", &IntegrationConfig, t)
}

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(10 * time.Second)
	StartChrome()
})

var _ = AfterSuite(func() {
	StopWebdriver()
})
