package mariadb_start_manager

import (
	"../os_helper"
	"fmt"
	"regexp"
)

type MariaDBStartManager struct {
	osHelper          os_helper.OsHelper
	logFileLocation   string
	stateFileLocation string
	mysqlServerPath   string
	username          string
	password          string
	jobIndex          int
}

func New(osHelper os_helper.OsHelper,
	logFileLocation string,
	stateFileLocation string,
	mysqlServerPath string,
	username string,
	password string,
	jobIndex int) *MariaDBStartManager {
	return &MariaDBStartManager{
		osHelper:          osHelper,
		logFileLocation:   logFileLocation,
		stateFileLocation: stateFileLocation,
		username:          username,
		password:          password,
		jobIndex:          jobIndex,
		mysqlServerPath:   mysqlServerPath,
	}
}

func (m *MariaDBStartManager) Execute() {
	if m.jobIndex == 0 {
		if m.osHelper.FileExists(m.stateFileLocation) {
			orig_contents, _ := m.osHelper.ReadFile(m.stateFileLocation)
			fmt.Printf("file exists and contains: '%s'\n", orig_contents)

			if orig_contents == "BOOTSTRAP" {
				fmt.Printf("starting in bootstrap")
				err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "bootstrap")
				if err != nil {
					panic(err)
				}

				m.osHelper.WriteStringToFile(m.stateFileLocation, "JOIN")
			} else {
				m.joinCluster()
			}
		} else {
			fmt.Printf("file does not exist, creating with contents: BOOTSTRAP\n")
			m.osHelper.WriteStringToFile(m.stateFileLocation, "BOOTSTRAP")

			fmt.Printf("starting in bootstrap \n")
			err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "bootstrap")
			if err != nil {
				panic(err)
			}

			m.upgradeAndRestartIfNecessary()
		}
	} else {
		m.joinCluster()
	}
}

func (m *MariaDBStartManager) joinCluster() {
	fmt.Printf("starting in join\n")
	err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "start")
	if err != nil {
		panic(err)
	}

	m.upgradeAndRestartIfNecessary()

	fmt.Printf("updating file with contents: JOIN\n")
	m.osHelper.WriteStringToFile(m.stateFileLocation, "JOIN")
}

func (m *MariaDBStartManager) upgradeAndRestartIfNecessary() {
	fmt.Printf("performing upgrade\n")
	output, err := m.osHelper.RunCommand(
		"bash",
		"mysql_upgrade.sh",
		m.username,
		m.password,
		m.logFileLocation)

	if m.requiresRestart(output, err) {
		fmt.Printf("stopping mysql\n")
		err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "stop")
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Printf("updating file with contents: JOIN\n")
		m.osHelper.WriteStringToFile(m.stateFileLocation, "JOIN")
	}
}

func (m *MariaDBStartManager) requiresRestart(output string, err error) bool {
	// No error indicates that the upgrade script performed an upgrade.
	if err == nil {
		fmt.Printf("upgrade sucessful - restart required\n")
		return true
	}
	fmt.Printf("upgrade output: %s\n", output)

	//known error messages where a restart should not occur, do not remove from
	acceptableErrorsCompiled, _ := regexp.Compile("already upgraded|Unknown command|WSREP has not yet prepared node")
	if acceptableErrorsCompiled.MatchString(output) {
		fmt.Printf("output string matches acceptable errors - skip restart\n")
		return false
	} else {
		fmt.Printf("output string does not match acceptable errors - restart required\n")
		return true
	}
}
