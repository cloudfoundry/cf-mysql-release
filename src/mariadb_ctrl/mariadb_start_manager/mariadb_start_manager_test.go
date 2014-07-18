package mariadb_start_manager_test

import (

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	manager "."
	"../os_helper/fakes"
	"errors"
)


var _ =	Describe("MariadbStartManager", func() {

	var mgr *manager.MariaDBStartManager
	var fake *fakes.FakeOsHelper

	logFileLocation := "/some-unused-location"
	stateFileLocation := "/another-unused-location"
	username := "fake-username"
	password := "fake-password"

	Describe("Execute on node >0", func(){
		BeforeEach(func() {

			fake = new(fakes.FakeOsHelper)

			mgr = manager.New(
				fake,
				logFileLocation,
				stateFileLocation,
				username,
				password,
				1)
		})

		Context("When the node needs to restart after upgrade", func(){
			It("Should start up in join mode, writes JOIN to a file, runs upgrade, stops mysql", func(){
				mgr.Execute()

				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(2))
				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_join.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

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

				executable, args = fake.RunCommandPanicOnErrArgsForCall(1)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_stop.sh"))
				Expect(args[1]).To(Equal(logFileLocation))
			})
		})

		Context("When the node does NOT need to restart after upgrade", func(){
			It("Should start up in join mode, writes JOIN to a file, runs upgrade", func(){
				fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
					return "This installation of MySQL is already upgraded to 10.0.12-MariaDB, use --force if you still need to run mysql_upgrade", errors.New("unused error text")
				}

				mgr.Execute()

				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(1))
				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_join.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

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
		})
	})

	Describe("Execute on node 0", func() {

		BeforeEach(func() {

			fake = new(fakes.FakeOsHelper)

			mgr = manager.New(
				fake,
				logFileLocation,
				stateFileLocation,
				username,
				password,
				0)
		})
		Context("When file is not present on node 0 and upgrade requires restart", func() {
			It("Should boostrap, upgrade and restart", func() {
				fake.FileExistsReturns(false)

				mgr.Execute()

				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("BOOTSTRAP"))

				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(2))
				Expect(fake.RunCommandCallCount()).To(Equal(1))

				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_bootstrap.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))

				executable, args = fake.RunCommandPanicOnErrArgsForCall(1)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_stop.sh"))
				Expect(args[1]).To(Equal(logFileLocation))
			})
		})
		Context("When file is not present and upgrade does not require restart", func() {
			It("Should boostrap, upgrade and write JOIN to file", func() {
				fake.FileExistsReturns(false)
				fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
					return "already upgraded", errors.New("unused error text")
				}

				mgr.Execute()

				Expect(fake.WriteStringToFileCallCount()).To(Equal(2))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("BOOTSTRAP"))

				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(1))
				Expect(fake.RunCommandCallCount()).To(Equal(1))

				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_bootstrap.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

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
		})
		Context("When file is present and reads 'BOOTSTRAP'", func() {
			It("Should bootstrap, and not upgrade", func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("BOOTSTRAP", nil)

				mgr.Execute()

				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(1))
				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_bootstrap.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))
			})
		})
		Context("When file is present and reads 'JOIN', and upgrade returns err: 'already upgraded'", func() {
			It("Should join, perform upgrade and not restart", func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("JOIN", nil)

				fake.RunCommandStub = func(arg1 string, arg2 ...string) (string, error) {
					return "already upgraded", errors.New("unused error text")
				}

				mgr.Execute()

				Expect(fake.RunCommandCallCount()).To(Equal(1))
				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(1))

				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_join.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))
			})
		})
		Context("When file is present and reads 'JOIN', and upgrade requires restart", func() {
			It("Should join, perform upgrade and restart", func() {
				fake.FileExistsReturns(true)
				fake.ReadFileReturns("JOIN", nil)

				mgr.Execute()
				Expect(fake.WriteStringToFileCallCount()).To(Equal(1))
				filename, contents := fake.WriteStringToFileArgsForCall(0)
				Expect(filename).To(Equal(stateFileLocation))
				Expect(contents).To(Equal("JOIN"))

				Expect(fake.RunCommandPanicOnErrCallCount()).To(Equal(2))
				Expect(fake.RunCommandCallCount()).To(Equal(1))

				executable, args := fake.RunCommandPanicOnErrArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_join.sh"))
				Expect(args[1]).To(Equal(logFileLocation))

				executable, args = fake.RunCommandArgsForCall(0)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_upgrade.sh"))
				Expect(args[1]).To(Equal(username))
				Expect(args[2]).To(Equal(password))
				Expect(args[3]).To(Equal(logFileLocation))

				executable, args = fake.RunCommandPanicOnErrArgsForCall(1)
				Expect(executable).To(Equal("bash"))
				Expect(args[0]).To(Equal("mysql_stop.sh"))
				Expect(args[1]).To(Equal(logFileLocation))
			})
		})
    })
})
