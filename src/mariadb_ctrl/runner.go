package main

import (
	"flag"
	"./os_helper"
	manager "./mariadb_start_manager"
)

var logFileLocation = flag.String(
	"logFile",
	"",
	"Specifies the location of the log file mysql sends logs to",
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

func main() {
	flag.Parse()

	mgr := manager.New(os_helper.NewImpl(),
		*logFileLocation,
		*stateFileLocation,
		*mysqlUser,
		*mysqlPassword,
		*jobIndex)
	mgr.Execute()
}
