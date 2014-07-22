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

	ensureMySQLCommandsRanWithOptions := func (options []string) {
		Expect(fake.RunCommandWithTimeoutCallCount()).To(Equal(len(options)))
		for i, option := range(options) {
			timeout, logFile, executable, args := fake.RunCommandWithTimeoutArgsForCall(i)
			Expect(timeout).To(Equal(300))
			Expect(logFile).To(Equal("/some-unused-location"))
			Expect(executable).To(Equal("bash"))
			Expect(args).To(Equal([]string{"/some-server-location", option}))
		}
	}

	ensureUpgrade := func () {
		executable, args := fake.RunCommandArgsForCall(0)
		Expect(executable).To(Equal("bash"))
		Expect(args[0]).To(Equal("mysql_upgrade.sh"))
		Expect(args[1]).To(Equal(username))
		Expect(args[2]).To(Equal(password))
		Expect(args[3]).To(Equal(logFileLocation))
	}

	ensureStateFileContentIs := func (expected string) {
		count := fake.WriteStringToFileCallCount()
		filename, contents := fake.WriteStringToFileArgsForCall(count-1)
		Expect(filename).To(Equal(stateFileLocation))
		Expect(contents).To(Equal(expected))
	}

	fakeRestartNOTNeededAfterUpgrade := func () {
		fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
			return "This installation of MySQL is already upgraded to 10.0.12-MariaDB, use --force if you still need to run mysql_upgrade",
			errors.New("unused error text")
		}
	}
		
	Describe("When starting in single-node deployment", func() {

		BeforeEach(func() {
			fake = new(fakes.FakeOsHelper)

			mgr = manager.New(
				fake,
				logFileLocation,
				stateFileLocation,
				mysqlServerPath,
				username,
				password,
				0, 1, false)
		})

		Context("On initial deploy, when it needs to be restarted after upgrade", func() {
			It("Starts in bootstrap mode", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"bootstrap","stop"})
				ensureStateFileContentIs("SINGLE_NODE")
				ensureUpgrade()
			})
		})

		Context("When a restart after upgrade is not necessary", func() {
			BeforeEach(func() {
				fakeRestartNOTNeededAfterUpgrade()
			})

			It("Starts in bootstrap mode", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"bootstrap"})
				ensureStateFileContentIs("SINGLE_NODE")
				ensureUpgrade()
			})
		})

		Context("When redeploying, and a restart after upgrade is necessary", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("SINGLE_NODE", nil)
			})
			It("Starts in bootstrap mode", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"bootstrap","stop"})
				ensureStateFileContentIs("SINGLE_NODE")
				ensureUpgrade()
			})
		})

	})

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
				1, 3, false)
		})

		Context("When the node needs to restart after upgrade", func() {
			It("Should start up in join mode, writes JOIN to a file, runs upgrade, stops mysql", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"start","stop"})
				ensureStateFileContentIs("JOIN")
				ensureUpgrade()
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
				fakeRestartNOTNeededAfterUpgrade()
			})
			It("Should start up in join mode, writes JOIN to a file, runs upgrade", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"start"})
				ensureStateFileContentIs("JOIN")
				ensureUpgrade()
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
				0, 3, false)
		})

		Context("When file is not present on node 0 and upgrade requires restart", func() {
			BeforeEach(func() {
				fake.FileExistsReturns(false)
			})
			It("Should boostrap, upgrade and restart", func() {
				mgr.Execute()
				ensureStateFileContentIs("BOOTSTRAP")
				ensureMySQLCommandsRanWithOptions([]string{"bootstrap","stop"})
				ensureUpgrade()

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
				fakeRestartNOTNeededAfterUpgrade()
			})
			It("Should boostrap, upgrade and write JOIN to file", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"bootstrap"})
				ensureUpgrade()
				ensureStateFileContentIs("JOIN")
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
				ensureMySQLCommandsRanWithOptions([]string{"bootstrap"})
				ensureStateFileContentIs("JOIN")
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
				fakeRestartNOTNeededAfterUpgrade()
			})
			It("Should join, perform upgrade and not restart", func() {
				mgr.Execute()
				ensureMySQLCommandsRanWithOptions([]string{"start"})
				ensureUpgrade()
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
				ensureMySQLCommandsRanWithOptions([]string{"start","stop"})
				ensureStateFileContentIs("JOIN")
				ensureUpgrade()

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

	Describe("When scaling the cluster", func(){

		BeforeEach(func() {
			fake = new(fakes.FakeOsHelper)
		})

		Context("When scaling down from many nodes to single", func() {
			BeforeEach(func() {
				mgr = manager.New(
					fake,
					logFileLocation,
					stateFileLocation,
					mysqlServerPath,
					username,
					password,
					0, 1, false)

				fake.FileExistsReturns(true)
				fake.ReadFileReturns("JOIN", nil)
			})
			Context("When restart is needed after upgrade", func(){
				It("Bootstraps node zero and writes SINGLE_NODE to file", func(){
					mgr.Execute()
					ensureMySQLCommandsRanWithOptions([]string{"bootstrap","stop"})
					ensureStateFileContentIs("SINGLE_NODE")
					ensureUpgrade()
				})
			})
		    Context("When no restart is needed", func(){
				BeforeEach(func() {
					fakeRestartNOTNeededAfterUpgrade()
				})

				It("Bootstraps node zero and writes SINGLE_NODE to file", func(){
					mgr.Execute()
					ensureMySQLCommandsRanWithOptions([]string{"bootstrap"})
					ensureStateFileContentIs("SINGLE_NODE")
					ensureUpgrade()
				})
			})
		})

		Context("Scaling from one to many nodes", func() {
			BeforeEach(func() {
				mgr = manager.New(
					fake,
					logFileLocation,
					stateFileLocation,
					mysqlServerPath,
					username,
					password,
					0, 3, false)

				fake.FileExistsReturns(true)
				fake.ReadFileReturns("SINGLE_NODE", nil)
			})
			Context("When a restart after upgrade is necessary", func() {
				It("bootstraps the first node and writes BOOTSTRAP to file", func() {
					mgr.Execute()
					ensureMySQLCommandsRanWithOptions([]string{"bootstrap","stop"})
					ensureStateFileContentIs("BOOTSTRAP")
					ensureUpgrade()
				})
			})
			Context("When a restart after upgrade is NOT necessary", func() {
				BeforeEach(func() {
					fakeRestartNOTNeededAfterUpgrade()
				})
				It("bootstraps the first node and writes JOIN to file", func() {
					mgr.Execute()
					ensureMySQLCommandsRanWithOptions([]string{"bootstrap"})
					ensureStateFileContentIs("JOIN")
					ensureUpgrade()
				})
			})
		})
	})
})
