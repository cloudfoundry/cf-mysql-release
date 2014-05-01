package cf_mysql_service

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"fmt"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/generator"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/runner"
	"time"
)

var (
	_ = Describe("P-MySQL Service", func() {
		serviceName := "p-mysql"
		planName := "100mb"
		timeout := 120.0
		retryInterval := 1.0

		It("Registers a route", func() {
			uri := SystemUri("p-mysql") + "/v2/catalog"

			fmt.Println("Curling url: ", uri)
			Eventually(Curl(uri), timeout, retryInterval).Should(Say("HTTP Basic: Access denied."))
		})

		Describe("Service instance lifecycle", func() {
			var appName string

			BeforeEach(func() {
				appName = RandomName()

				Eventually(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), 60*time.Second).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(Cf("delete", appName, "-f"), 20*time.Second).Should(Exit(0))
			})

			Context("using a new service instance", func() {
				It("Allows users to create, bind, write to, read from, unbind, and destroy the service instance", func() {
					serviceInstanceName := RandomName()
					uri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"

					Eventually(Cf("create-service", serviceName, planName, serviceInstanceName), 60*time.Second).Should(Exit(0))
					Eventually(Cf("bind-service", appName, serviceInstanceName), 60*time.Second).Should(Exit(0))
					Eventually(Cf("start", appName), 5*60*time.Second).Should(Exit(0))

					fmt.Println("Posting to url: ", uri)
					Eventually(Curl("-d", "myvalue", uri), timeout, retryInterval).Should(Say("myvalue"))
					fmt.Println("\n")

					fmt.Println("Curling url: ", uri)
					Eventually(Curl(uri), timeout, retryInterval).Should(Say("myvalue"))
					fmt.Println("\n")

					Eventually(Cf("unbind-service", appName, serviceInstanceName), 20*time.Second).Should(Exit(0))
					Eventually(Cf("delete-service", "-f", serviceInstanceName), 20*time.Second).Should(Exit(0))
				})
			})
		})

		Describe("Enforcing MySQL quota", func() {
			var appName string
			var serviceInstanceName string

			BeforeEach(func() {
				appName = RandomName()
				serviceInstanceName = RandomName()

				Eventually(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), 60*time.Second).Should(Exit(0))
				Eventually(Cf("create-service", serviceName, planName, serviceInstanceName), 60*time.Second).Should(Exit(0))
				Eventually(Cf("bind-service", appName, serviceInstanceName), 60*time.Second).Should(Exit(0))
				Eventually(Cf("start", appName), 5*60*time.Second).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(Cf("unbind-service", appName, serviceInstanceName), 20*time.Second).Should(Exit(0))
				Eventually(Cf("delete-service", "-f", serviceInstanceName), 20*time.Second).Should(Exit(0))
				Eventually(Cf("delete", appName, "-f"), 20*time.Second).Should(Exit(0))
			})

			It("enforces the storage quota", func() {
				quotaEnforcerSleepTime := 10 * time.Second
				uri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"
				writeUri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/write-bulk-data"
				deleteUri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/delete-bulk-data"
				firstValue := RandomName()[:20]
				secondValue := RandomName()[:20]

				fmt.Println("*** Proving we can write")
				Eventually(Curl("-d", firstValue, uri), timeout, retryInterval).Should(Say(firstValue))
				fmt.Println("*** Proving we can read")
				Eventually(Curl(uri), timeout, retryInterval).Should(Say(firstValue))

				fmt.Println("*** Exceeding quota")
				Eventually(Curl("-d", "100", writeUri), timeout, retryInterval).Should(Say("Database now contains"))

				fmt.Println("*** Sleeping to let quota enforcer run")
				time.Sleep(quotaEnforcerSleepTime)

				fmt.Println("*** Proving we cannot write")
				Eventually(Curl("-d", firstValue, uri), timeout, retryInterval).Should(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
				fmt.Println("*** Proving we can read")
				Eventually(Curl(uri), timeout, retryInterval).Should(Say(firstValue))

				fmt.Println("*** Deleting below quota")
				Eventually(Curl("-d", "20", deleteUri), timeout, retryInterval).Should(Say("Database now contains"))

				fmt.Println("*** Sleeping to let quota enforcer run")
				time.Sleep(quotaEnforcerSleepTime)

				fmt.Println("*** Proving we can write")
				Eventually(Curl("-d", secondValue, uri), timeout, retryInterval).Should(Say(secondValue))
				fmt.Println("*** Proving we can read")
				Eventually(Curl(uri), timeout, retryInterval).Should(Say(secondValue))
			})
		})
	})
)
