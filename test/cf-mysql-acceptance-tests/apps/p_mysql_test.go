package apps

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/vito/cmdtest/matchers"

	"fmt"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/cf"
	. "github.com/pivotal-cf-experimental/cf-test-helpers/generator"
	"time"
)

var (
	_ = Describe("P-MySQL Service", func() {
		serviceName := "p-mysql"
		planName := "100mb"
		timeout := 10.0
		retryInterval := 1.0

		It("Registers a route", func() {
			uri := AppUri("pmysql") + "/v2/catalog"

			fmt.Println("Curling url: ", uri)
			Eventually(Curling(uri), timeout, retryInterval).Should(Say("HTTP Basic: Access denied."))
		})

		Describe("Service instance lifecycle", func() {
			var appName string

			BeforeEach(func() {
				appName = RandomName()

				Expect(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start")).To(ExitWithTimeout(0, 60*time.Second))
			})

			AfterEach(func() {
				Expect(Cf("delete", appName, "-f")).To(ExitWithTimeout(0, 20*time.Second))
			})

			Context("using a new service instance", func() {
				It("Allows users to create, bind, write to, read from, unbind, and destroy the service instance", func() {
					serviceInstanceName := RandomName()
					uri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"

					Expect(Cf("create-service", serviceName, planName, serviceInstanceName)).To(ExitWithTimeout(0, 60*time.Second))
					Expect(Cf("bind-service", appName, serviceInstanceName)).To(ExitWithTimeout(0, 60*time.Second))
					Expect(Cf("start", appName)).To(ExitWithTimeout(0, 5*60*time.Second))

					fmt.Println("Posting to url: ", uri)
					Eventually(Curling("-d", "myvalue", uri), timeout, retryInterval).Should(Say("myvalue"))
					fmt.Println("\n")

					fmt.Println("Curling url: ", uri)
					Eventually(Curling(uri), timeout, retryInterval).Should(Say("myvalue"))
					fmt.Println("\n")

					Expect(Cf("unbind-service", appName, serviceInstanceName)).To(ExitWithTimeout(0, 20*time.Second))
					Expect(Cf("delete-service", "-f", serviceInstanceName)).To(ExitWithTimeout(0, 20*time.Second))
				})
			})

			Context("using a long-running service instance", func() {
				It("Allows users to bind, write to, read from, and unbind the service instance", func() {
					serviceInstanceName := IntegrationConfig.ExistingServiceInstanceName
					testValue := RandomName()[:20]
					uri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"

					Expect(Cf("bind-service", appName, serviceInstanceName)).To(ExitWithTimeout(0, 60*time.Second))
					Expect(Cf("start", appName)).To(ExitWithTimeout(0, 5*60*time.Second))

					fmt.Println("Posting to url: ", uri)
					Eventually(Curling("-d", testValue, uri), timeout, retryInterval).Should(Say(testValue))
					fmt.Println("\n")

					fmt.Println("Curling url: ", uri)
					Eventually(Curling(uri), timeout, retryInterval).Should(Say(testValue))
					fmt.Println("\n")

					Expect(Cf("unbind-service", appName, serviceInstanceName)).To(ExitWithTimeout(0, 20*time.Second))
				})
			})
		})

		Describe("Enforcing MySQL quota", func() {
			var appName string
			var serviceInstanceName string

			BeforeEach(func() {
				appName = RandomName()
				serviceInstanceName = RandomName()

				Expect(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start")).To(ExitWithTimeout(0, 60*time.Second))
				Expect(Cf("create-service", serviceName, planName, serviceInstanceName)).To(ExitWithTimeout(0, 60*time.Second))
				Expect(Cf("bind-service", appName, serviceInstanceName)).To(ExitWithTimeout(0, 60*time.Second))
				Expect(Cf("start", appName)).To(ExitWithTimeout(0, 5*60*time.Second))
			})

			AfterEach(func() {
				Expect(Cf("unbind-service", appName, serviceInstanceName)).To(ExitWithTimeout(0, 20*time.Second))
				Expect(Cf("delete-service", "-f", serviceInstanceName)).To(ExitWithTimeout(0, 20*time.Second))
				Expect(Cf("delete", appName, "-f")).To(ExitWithTimeout(0, 20*time.Second))
			})

			It("enforces the storage quota", func() {
				quotaEnforcerSleepTime := 2 * time.Second
				uri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/mykey"
				writeUri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/write-bulk-data"
				deleteUri := AppUri(appName) + "/service/mysql/" + serviceInstanceName + "/delete-bulk-data"
				firstValue := RandomName()[:20]
				secondValue := RandomName()[:20]

				fmt.Println("*** Proving we can write")
				Eventually(Curling("-d", firstValue, uri), timeout, retryInterval).Should(Say(firstValue))
				fmt.Println("*** Proving we can read")
				Eventually(Curling(uri), timeout, retryInterval).Should(Say(firstValue))

				fmt.Println("*** Exceeding quota")
				Eventually(Curling("-d", "100", writeUri), timeout, retryInterval).Should(Say("Database now contains"))

				fmt.Println("*** Sleeping to let quota enforcer run")
				time.Sleep(quotaEnforcerSleepTime)

				fmt.Println("*** Proving we cannot write")
				Eventually(Curling("-d", firstValue, uri), timeout, retryInterval).Should(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
				fmt.Println("*** Proving we can read")
				Eventually(Curling(uri), timeout, retryInterval).Should(Say(firstValue))

				fmt.Println("*** Deleting below quota")
				Eventually(Curling("-d", "20", deleteUri), timeout, retryInterval).Should(Say("Database now contains"))

				fmt.Println("*** Sleeping to let quota enforcer run")
				time.Sleep(quotaEnforcerSleepTime)

				fmt.Println("*** Proving we can write")
				Eventually(Curling("-d", secondValue, uri), timeout, retryInterval).Should(Say(secondValue))
				fmt.Println("*** Proving we can read")
				Eventually(Curling(uri), timeout, retryInterval).Should(Say(secondValue))
			})
		})
	})
)
