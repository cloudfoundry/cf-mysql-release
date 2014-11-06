package dashboard_test

import (
	"../../helpers"
	"testing"
)

var IntegrationConfig = helpers.LoadConfig()

func TestDashboard(t *testing.T) {
	helpers.PrepareAndRunTests("Dashboard", &IntegrationConfig, t)
}
