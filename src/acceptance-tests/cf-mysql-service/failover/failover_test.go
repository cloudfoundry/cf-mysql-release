package failover_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/sclevine/agouti/dsl"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	"../../partition"

	context_setup "github.com/cloudfoundry-incubator/cf-test-helpers/services/context_setup"
)

const (
	firstKey    = "mykey"
	firstValue  = "myvalue"
	secondKey   = "mysecondkey"
	secondValue = "mysecondvalue"
	planName    = "100mb-dev"

	sinatraPath = "../../assets/sinatra_app"
)

var appName string
var minuteTimeout, curlTimeout, retryInterval time.Duration

func createAndBindService(serviceName, serviceInstanceName, planName string) {
	minuteTimeout := context_setup.ScaledTimeout(60 * time.Second)
	retryInterval := 10 * time.Second
	fmt.Print()
	Eventually(func() *Session {
		session := Cf("create-service", serviceName, planName, serviceInstanceName)
		session.Wait(minuteTimeout)
		return session
	}, minuteTimeout*3, retryInterval).Should(Exit(0))

	Eventually(func() *Session {
		session := Cf("bind-service", appName, serviceInstanceName)
		session.Wait(minuteTimeout)
		return session
	}, minuteTimeout*3, retryInterval).Should(Exit(0))

	Eventually(func() *Session {
		session := Cf("restart", appName)
		session.Wait(minuteTimeout)
		return session
	}, minuteTimeout*3, minuteTimeout/2).Should(Exit(0))
}

func assertAppIsRunning(appName string) {
	pingUri := AppUri(appName) + "/ping"
	Eventually(Curl(pingUri), curlTimeout, retryInterval).Should(Say("OK"))
}

func assertWriteToDB(key, value, uri string) {
	writeTimeout := context_setup.ScaledTimeout(60*time.Second) * 5
	retryInterval := 10 * time.Second
	curlURI := uri + "/" + key
	Eventually(Curl("-d", value, curlURI), writeTimeout, retryInterval).Should(Say(value))
}

func assertReadFromDB(key, value, uri string) {
	readTimeout := context_setup.ScaledTimeout(60*time.Second) * 2
	retryInterval := 10 * time.Second
	curlURI := uri + "/" + key
	Eventually(Curl(curlURI), readTimeout, retryInterval).Should(Say(value))
}

var _ = Feature("CF MySQL Failover", func() {
	BeforeEach(func() {
		appName = generator.RandomName()
		minuteTimeout = context_setup.ScaledTimeout(60 * time.Second)
		curlTimeout = minuteTimeout * 2
		retryInterval = 10 * time.Second

		Step("Push an app", func() {
			Eventually(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), minuteTimeout, retryInterval).Should(Exit(0))
		})
	})

	Context("when the mysql node is partitioned", func() {
		BeforeEach(func() {
			Expect(IntegrationConfig.MysqlNodes).NotTo(BeNil())
			Expect(len(IntegrationConfig.MysqlNodes)).To(BeNumerically(">=", 1))
		})

		AfterEach(func() {
			// Re-introducing a mariadb node once partitioned is unsafe
			// See https://www.pivotaltracker.com/story/show/81974864
			// partition.Off(IntegrationConfig.MysqlNodes[0].SshTunnel)
		})

		Scenario("write/read data before the partition and successfully writes and read it after", func() {
			planName := "100mb-dev"
			serviceInstanceName := generator.RandomName()
			instanceURI := AppUri(appName) + "/service/mysql/" + serviceInstanceName

			Step("Create & bind a DB", func() {
				createAndBindService(IntegrationConfig.ServiceName, serviceInstanceName, planName)
				assertAppIsRunning(appName)
			})

			Step("Start App", func() {
				Eventually(Cf("start", appName), minuteTimeout*5).Should(Exit(0))
				assertAppIsRunning(appName)
			})

			Step("Write a key-value pair to DB", func() {
				assertWriteToDB(firstKey, firstValue, instanceURI)
			})

			Step("Read value from DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
			})

			Step("Take down mysql node", func() {
				partition.On(
					IntegrationConfig.MysqlNodes[0].SshTunnel,
					IntegrationConfig.MysqlNodes[0].Ip,
				)
			})

			Step("Restart sinatra app to reset connections", func() {
				fmt.Println("Restarting app")
				session := Cf("restart", appName)
				timeout := make(chan bool, 1)
				go func() {
					time.Sleep(2 * time.Minute)
					timeout <- true
				}()
				select {
				case <-session.Exited:
					fmt.Println("App restarted")
				case <-timeout:
					fmt.Println("Restart timed out")
				}
				fmt.Println("Checking whether app is running")
				assertAppIsRunning(appName)
			})

			Step("Write a second key-value pair to DB", func() {
				assertWriteToDB(secondKey, secondValue, instanceURI)
			})

			Step("Read both values from DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
				assertReadFromDB(secondKey, secondValue, instanceURI)
			})
		})
	})

	Context("Broker failure", func() {
		var broker0SshTunnel, broker1SshTunnel string

		BeforeEach(func() {
			Expect(IntegrationConfig.Brokers).NotTo(BeNil())
			Expect(len(IntegrationConfig.Brokers)).To(BeNumerically(">=", 2))

			broker0SshTunnel = IntegrationConfig.Brokers[0].SshTunnel
			broker1SshTunnel = IntegrationConfig.Brokers[1].SshTunnel
		})

		AfterEach(func() {
			partition.Off(broker0SshTunnel)
			partition.Off(broker1SshTunnel)
		})

		Scenario("Broker failure", func() {
			serviceInstanceName := generator.RandomName()
			instanceURI := AppUri(appName) + "/service/mysql/" + serviceInstanceName

			// Remove partitions in case previous test did not cleanup correctly
			partition.Off(broker0SshTunnel)
			partition.Off(broker1SshTunnel)

			Step("Take down first broker instance", func() {
				partition.On(broker0SshTunnel, IntegrationConfig.Brokers[0].Ip)
			})

			Step("Create & bind a DB", func() {
				createAndBindService(IntegrationConfig.ServiceName, serviceInstanceName, planName)
			})

			Step("Write a key-value pair to DB", func() {
				assertWriteToDB(firstKey, firstValue, instanceURI)
			})

			Step("Read valuefrom DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
			})

			Step("Bring back first broker instance", func() {
				partition.Off(broker0SshTunnel)
			})

			Step("Take down second broker instance", func() {
				partition.On(broker1SshTunnel, IntegrationConfig.Brokers[1].Ip)
			})

			Step("Create & bind a DB again", func() {
				serviceInstanceName := generator.RandomName()
				createAndBindService(IntegrationConfig.ServiceName, serviceInstanceName, planName)
			})

			Step("Write a second key-value pair to DB", func() {
				assertWriteToDB(secondKey, secondValue, instanceURI)
			})

			Step("Read both values from DB", func() {
				assertReadFromDB(firstKey, firstValue, instanceURI)
				assertReadFromDB(secondKey, secondValue, instanceURI)
			})

			Step("Bring back second broker instance", func() {
				partition.Off(broker1SshTunnel)
			})
		})
	})
})
