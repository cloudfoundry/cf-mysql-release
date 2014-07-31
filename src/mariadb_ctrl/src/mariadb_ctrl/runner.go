package main

import (
	"flag"

	manager "mariadb_ctrl/mariadb_start_manager"
	"mariadb_ctrl/os_helper"
)

var logFileLocation = flag.String(
	"logFile",
	"",
	"Specifies the location of the log file mysql sends logs to",
)

var mysqlServerPath = flag.String(
	"mysqlServer",
	"",
	"Specifies the location of the mysql.server file",
)

var dbSeedScriptPath = flag.String(
	"dbSeedScript",
	"",
	"Specifies the location of the script that seeds the server with databases",
)

var stateFileLocation = flag.String(
	"stateFile",
	"",
	"Specifies the location to store the statefile for MySQL boot",
)

var mysqlUser = flag.String(
	"mysqlUser",
	"root",
	"Specifies the user name for MySQL",
)

var mysqlPassword = flag.String(
	"mysqlPassword",
	"",
	"Specifies the password for connecting to MySQL",
)

var jobIndex = flag.Int(
	"jobIndex",
	1,
	"Specifies the job index of the MySQL node",
)

var numberOfNodes = flag.Int(
	"numberOfNodes",
	3,
	"Number of nodes deployed in the galera cluster",
)

func main() {
	flag.Parse()

	mgr := manager.New(os_helper.NewImpl(),
		*logFileLocation,
		*stateFileLocation,
		*mysqlServerPath,
		*mysqlUser,
		*mysqlPassword,
		*dbSeedScriptPath,
		*jobIndex,
		*numberOfNodes,
		true,
	)
	mgr.Execute()
}
