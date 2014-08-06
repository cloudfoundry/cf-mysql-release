package mariadb_start_manager

import (
	"fmt"
	"mariadb_ctrl/os_helper"
	"mariadb_ctrl/galera_helper"
	"regexp"
	"time"
)

type MariaDBStartManager struct {
	osHelper               os_helper.OsHelper
	logFileLocation        string
	stateFileLocation      string
	mysqlServerPath        string
	username               string
	password               string
	jobIndex               int
	numberOfNodes          int
	loggingOn              bool
	dbSeedScriptPath       string
	upgradeScriptPath      string
	mysqlCommandScriptPath string
	ClusterReachabilityChecker galera_helper.ClusterReachabilityChecker
}

func New(osHelper os_helper.OsHelper,
	logFileLocation string,
	stateFileLocation string,
	mysqlServerPath string,
	username string,
	password string,
	dbSeedScriptPath string,
	jobIndex int,
	numberOfNodes int,
	loggingOn bool,
	upgradeScriptPath string,
	mysqlCommandScriptPath string,
	clusterReachabilityChecker galera_helper.ClusterReachabilityChecker,
) *MariaDBStartManager {
	return &MariaDBStartManager{
		osHelper:               osHelper,
		logFileLocation:        logFileLocation,
		stateFileLocation:      stateFileLocation,
		username:               username,
		password:               password,
		jobIndex:               jobIndex,
		mysqlServerPath:        mysqlServerPath,
		numberOfNodes:          numberOfNodes,
		loggingOn:              loggingOn,
		dbSeedScriptPath:       dbSeedScriptPath,
		upgradeScriptPath:      upgradeScriptPath,
		mysqlCommandScriptPath: mysqlCommandScriptPath,
		ClusterReachabilityChecker: clusterReachabilityChecker,
	}
}

func (m *MariaDBStartManager) Log(info string) {
	if m.loggingOn {
		fmt.Printf("%v ----- %v", time.Now().Local(), info)
	}
}

func (m *MariaDBStartManager) Execute() {
	//We should NEVER bootstrap unless we are Index 0
	if m.jobIndex == 0 {

		//single-node deploy
		if m.numberOfNodes == 1 {
			m.Log("Single node deploy")
			m.bootstrapUpgradeAndWriteState("SINGLE_NODE")
			return
		}

		//MULTI-NODE DEPLOYMENTS BELOW

		//intial deploy, state file does not exists
		if !m.osHelper.FileExists(m.stateFileLocation) {
			m.Log("state file does not exist, creating with contents: JOIN\n")
			m.bootstrapUpgradeAndWriteState("JOIN")
			return
		}

		//state file exists
		orig_contents, _ := m.osHelper.ReadFile(m.stateFileLocation)
		m.Log(fmt.Sprintf("state file exists and contains: '%s'\n", orig_contents))

		//scaling up from a single node cluster
		if orig_contents == "SINGLE_NODE" {
			m.bootstrapUpgradeAndWriteState("JOIN")
			return
		}
	}

	m.joinCluster()
}

func (m *MariaDBStartManager) bootstrapUpgradeAndWriteState(state string) {
	m.bootstrapAndWriteState(state)
	m.seedDatabases()
	m.upgradeAndRestartIfNecessary("bootstrap")
}

func (m *MariaDBStartManager) bootstrapAndWriteState(state string) {
	m.Log("updating file with contents: " + state + "\n")
	m.osHelper.WriteStringToFile(m.stateFileLocation, state)

	var mode string

	if m.ClusterReachabilityChecker.AnyNodesReachable() {
		mode = "start"
		m.Log("STARTING NODE IN JOIN MODE.\n")
	} else {
		mode = "bootstrap"
		m.Log("STARTING NODE IN BOOTSTRAPPING MODE.\n")
	}

	err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, mode)

	if err != nil {
		panic(err)
	}
}

func (m *MariaDBStartManager) joinCluster() {
	m.Log("STARTING NODE IN JOIN MODE.\n")
	err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "start")
	if err != nil {
		panic(err)
	}

	m.seedDatabases()
	m.upgradeAndRestartIfNecessary("start")

	m.Log("updating file with contents: JOIN\n")
	m.osHelper.WriteStringToFile(m.stateFileLocation, "JOIN")
}

func (m *MariaDBStartManager) seedDatabases() {
	output, err := m.osHelper.RunCommand("bash", m.dbSeedScriptPath)
	if err != nil {
		m.Log("Seeding databases failed:\n")
		m.Log(output + "\n")

		m.Log("STOPPING NODE.\n")
		m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "stop")

		panic(err)
	}
}

func (m *MariaDBStartManager) upgradeAndRestartIfNecessary(mode string) {
	m.Log("performing upgrade\n")

	_, disableReplicationErr := m.osHelper.RunCommand(
		"bash",
		m.mysqlCommandScriptPath,
		"SET global wsrep_on='OFF'",
		m.username,
		m.password,
		m.logFileLocation)

	if disableReplicationErr != nil {
		panic(disableReplicationErr)
	}

	_, checkReplicationEnabledErr := m.osHelper.RunCommand(
		"bash",
		m.mysqlCommandScriptPath,
		"SHOW variables LIKE 'wsrep_on'",
		m.username,
		m.password,
		m.logFileLocation)

	if checkReplicationEnabledErr != nil {
		panic(checkReplicationEnabledErr)
	}

	upgradeOutput, upgradeErr := m.osHelper.RunCommand(
		"bash",
		m.upgradeScriptPath,
		m.username,
		m.password,
		m.logFileLocation)

	_, enableReplicationErr := m.osHelper.RunCommand(
		"bash",
		m.mysqlCommandScriptPath,
		"SET global wsrep_on='ON'",
		m.username,
		m.password,
		m.logFileLocation)

	if enableReplicationErr != nil {
		panic(enableReplicationErr)
	}

	_, checkReplicationDisabledErr := m.osHelper.RunCommand(
		"bash",
		m.mysqlCommandScriptPath,
		"SHOW variables LIKE 'wsrep_on'",
		m.username,
		m.password,
		m.logFileLocation)

	if checkReplicationDisabledErr != nil {
		panic(checkReplicationDisabledErr)
	}

	if m.requiresRestart(upgradeOutput, upgradeErr) {
		m.Log("STOPPING NODE.\n")
		err := m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, "stop")
		if err != nil {
			panic(err)
		}

		if mode == "bootstrap" && m.ClusterReachabilityChecker.AnyNodesReachable() {
			mode = "start"
		}

		m.Log("STARTING NODE IN" + mode + " MODE.\n")
		err = m.osHelper.RunCommandWithTimeout(300, m.logFileLocation, "bash", m.mysqlServerPath, mode)
		if err != nil {
			panic(err)
		}
	}
}

func (m *MariaDBStartManager) requiresRestart(output string, err error) bool {
	// No error indicates that the upgrade script performed an upgrade.
	if err == nil {
		m.Log("upgrade sucessful - restart required\n")
		return true
	}
	m.Log(fmt.Sprintf("upgrade output: %s\n", output))

	//known error messages where a restart should not occur, do not remove from
	acceptableErrorsCompiled, _ := regexp.Compile("already upgraded|Unknown command|WSREP has not yet prepared node")
	if acceptableErrorsCompiled.MatchString(output) {
		m.Log("output string matches acceptable errors - skip restart\n")
		return false
	} else {
		m.Log("output string does not match acceptable errors - restart required\n")
		return true
	}
}
