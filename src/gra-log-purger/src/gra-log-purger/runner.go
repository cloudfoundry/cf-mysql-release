package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var graLogDir = flag.String(
	"graLogDir",
	"",
	"Specifies the directory from which to purge GRA log files.",
)

var graLogDaysToKeep = flag.Int(
	"graLogDaysToKeep",
	60,
	"Specifies the maximum age of the GRA log files allowed.",
)

var pidfile = flag.String(
	"pidfile",
	"",
	"The location for the pidfile",
)

func main() {
	flag.Parse()

	err := ioutil.WriteFile(*pidfile, []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		panic(err)
	}

	for {
		out, err := runCommand("sh", "/var/vcap/jobs/mysql/bin/gra-log-purger.sh", *graLogDir, strconv.Itoa(*graLogDaysToKeep))
		if err != nil {
			LogErrorWithTimestamp(err)
		}
		LogWithTimestamp(out)
		LogWithTimestamp("Sleeping for one hour\n")
		time.Sleep(1 * time.Hour)
	}
}

// Runs command with stdout and stderr pipes connected to process
func runCommand(executable string, args ...string) (string, error) {
	cmd := exec.Command(executable, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

func LogWithTimestamp(format string, args ...interface{}) {
	fmt.Printf("[%s] - ", time.Now().Local())
	if nil == args {
		fmt.Printf(format)
	} else {
		fmt.Printf(format, args...)
	}
}

func LogErrorWithTimestamp(err error) {
	fmt.Fprintf(os.Stderr, "[%s] - ", time.Now().Local())
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
}
