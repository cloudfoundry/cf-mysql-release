package mariadb_start_manager_test

import (
	manager "."
	"../os_helper/fakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MariadbStartManager", func() {

	var mgr *manager.MariaDBStartManager
	var fake *fakes.FakeOsHelper

	logFileLocation := "/some-unused-location"
	mysqlServerPath := "/some-server-location"
	stateFileLocation := "/another-unused-location"
	username := "fake-username"
	password := "fake-password"

	Describe("Execute on node >0", func() {
		BeforeEach(func() {

			fake = new(fakes.FakeOsHelper)

			mgr = manager.New(
				fake,
				logFileLocation,
				stateFileLocation,
				mysqlServerPath,
				username,
				password,
				1)
		})

		Context("When the node needs to restart after upgrade", func() {
			It("Should start up in join mode, writes JOIN to a file, runs upgrade, stops mysql", func() {
				mgr.Execute()

				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(2))
				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"start"}))

				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))

				timeout, logFile, executable, args = fake.RunCommandWithTimeoutArgsForCall(1)
				Expect(timeout).To(Equal(300))
				Expect(executable).To(Equal("bash"))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"stop",
				}))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
			Context("When stopping mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						if arg3[1] == "stop" {
							return errors.New("some errors")
						} else {
							return nil
						}
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})

		Context("When the node does NOT need to restart after upgrade", func() {
			BeforeEach(func() {
				fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
					return "This installation of MySQL is already upgraded to 10.0.12-MariaDB, use --force if you still need to run mysql_upgrade",
							errors.New("unused error text")
				}
			})
			It("Should start up in join mode, writes JOIN to a file, runs upgrade", func() {
				mgr.Execute()

				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(1))
				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"start"}))

				Expect(fake.WriteStringToFileCallCount()).To(BeNumerically(">=", 1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})
	})

	Describe("Execute on node 0", func() {

		BeforeEach(func() {

			fake = new(fakes.FakeOsHelper)

			mgr = manager.New(
				fake,
				logFileLocation,
				stateFileLocation,
				mysqlServerPath,
				username,
				password,
				0)
		})
		Context("When file is not present on node 0 and upgrade requires restart", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(false)
			})
			It("Should boostrap, upgrade and restart", func() {
				mgr.Execute()

				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("BOOTSTRAP"))

				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(2))
				Expect(fake.RunCommandCallCount()).To(Equal(1))

				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"bootstrap"}))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))

				timeout, logFile, executable, args = fake.RunCommandWithTimeoutArgsForCall(1)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"stop"}))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})
		Context("When file is not present and upgrade does not require restart", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(false)
				fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
					return "already upgraded", errors.New("unused error text")
				}
			})
			It("Should boostrap, upgrade and write JOIN to file", func() {
				mgr.Execute()

				Expect(fake.WriteStringToFileCallCount()).To(Equal(2))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("BOOTSTRAP"))

				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(1))
				Expect(fake.RunCommandCallCount()).To(Equal(1))

				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"bootstrap"}))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))

				filename, contents = fake.WriteStringToFileArgsForCall(1)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})
		Context("When file is present and reads 'BOOTSTRAP'", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("BOOTSTRAP", nil)
			})
			It("Should bootstrap, and not upgrade", func() {
				mgr.Execute()

				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(1))
				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"bootstrap"}))

				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})
		Context("When file is present and reads 'JOIN', and upgrade returns err: 'already upgraded'", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("JOIN", nil)
				fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
					return "already upgraded", errors.New("unused error text")
				}
			})
			It("Should join, perform upgrade and not restart", func() {
				mgr.Execute()

				Expect(fake.RunCommandCallCount()).To(Equal(1))
				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(1))

				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"start"}))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})
		Context("When file is present and reads 'JOIN', and upgrade requires restart", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("JOIN", nil)
			})
			It("Should join, perform upgrade and restart", func() {
				mgr.Execute()
				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))

				Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(2))
				Expect(fake.RunCommandCallCount()).To(Equal(1))

				timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(0)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"start"}))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))

				timeout, logFile, executable, args = fake.RunCommandWithTimeoutArgsForCall(1)
				Expect(timeout).To(Equal(300))
				Expect(logFile).To(Equal("/some-unused-location"))
				Expect(executable).To(Equal("bash"))
				Expect(args).To(Equal([]string{"/some-server-location",
					"stop"}))
			})
			Context("When starting mariadb causes an error", func() {
				It("Panics", func() {
					fake.RunCommandWithTimeoutStub = func(arg0 int, arg1 string, arg2 string, arg3 ...string) error {
						return errors.New("some error")
					}
					Expect(func() {
						mgr.Execute()
					}).To(Panic())
				})
			})
		})
	})
})
