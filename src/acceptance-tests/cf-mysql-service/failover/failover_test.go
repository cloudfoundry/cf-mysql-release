package failover_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	. "github.com/sclevine/agouti/dsl"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	"../../partition"

	context_setup "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

var _ = Feature("CF MySQL Failover", func() {
	var appName string
	var broker_0_ssh_tunnel, broker_1_ssh_tunnel, broker_0_local_ip, broker_1_local_ip string

	Background(func() {
		if IntegrationConfig.Brokers == nil || len(IntegrationConfig.Brokers) < 2 {
			panic(errors.New("Brokers not configured"))
		}

		broker_0_ssh_tunnel = IntegrationConfig.Brokers[0].SshTunnel
		broker_1_ssh_tunnel = IntegrationConfig.Brokers[1].SshTunnel
		broker_0_local_ip = IntegrationConfig.Brokers[0].Ip
		broker_1_local_ip = IntegrationConfig.Brokers[1].Ip

		const sinatraPath = "../../assets/sinatra_app"
		appName = generator.RandomName()

		Step("Push an app", func() {
			Eventually(
				Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"),
				context_setup.ScaledTimeout(60*time.Second),
			).Should(Exit(0))
		})
	})

	AfterEach(func() {
		partition.Off(broker_0_ssh_tunnel)
		partition.Off(broker_1_ssh_tunnel)
	})

	Scenario("Broker failure", func() {
		partition.Off(broker_0_ssh_tunnel)
		partition.Off(broker_1_ssh_tunnel)

		planName := "100mb-dev"

		Step("Take down first broker instance", func() {
			partition.On(broker_0_ssh_tunnel, broker_0_local_ip)
		})

		Step("Create & bind a DB", func() {
			serviceInstance1 := generator.RandomName()

			Eventually(func() *Session {
				session := Cf("create-service", IntegrationConfig.ServiceName, planName, serviceInstance1)
				session.Wait(context_setup.ScaledTimeout(60 * time.Second))
				return session
			}, context_setup.ScaledTimeout(60*time.Second)*3, 10*time.Second).Should(Exit(0))

			Eventually(func() *Session {
				session := Cf("bind-service", appName, serviceInstance1)
				session.Wait(context_setup.ScaledTimeout(60 * time.Second))
				return session
			}, context_setup.ScaledTimeout(60*time.Second)*3, 10*time.Second).Should(Exit(0))
		})

		Step("Bring back first broker instance", func() {
			partition.Off(broker_0_ssh_tunnel)
		})

		Step("Take down second broker instance", func() {
			partition.On(broker_1_ssh_tunnel, broker_1_local_ip)
		})

		Step("Create & bind a DB again", func() {
			serviceInstance2 := generator.RandomName()

			Eventually(func() *Session {
				session := Cf("create-service", IntegrationConfig.ServiceName, planName, serviceInstance2)
				session.Wait(context_setup.ScaledTimeout(60 * time.Second))
				return session
			}, context_setup.ScaledTimeout(60*time.Second)*3, 10*time.Second).Should(Exit(0))

			Eventually(func() *Session {
				session := Cf("bind-service", appName, serviceInstance2)
				session.Wait(context_setup.ScaledTimeout(60 * time.Second))
				return session
			}, context_setup.ScaledTimeout(60*time.Second)*3, 10*time.Second).Should(Exit(0))
		})
	})
})
