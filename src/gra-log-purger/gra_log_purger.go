package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"io/ioutil"
	"strings"
	"strconv"
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

	if *graLogDir == "" {
		logErrorWithTimestamp(fmt.Errorf("No gra log directory supplied"))
		os.Exit(1)
	}

	if *pidfile == "" {
		logErrorWithTimestamp(fmt.Errorf("No pidfile supplied"))
		os.Exit(1)
	}

	err := ioutil.WriteFile(*pidfile, []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		panic(err)
	}
	defer os.Remove(*pidfile)

	for {
		ageCutoff := time.Duration(*graLogDaysToKeep*24) * time.Hour
		logs, err := FindGraLogs(*graLogDir, ageCutoff)
		if err != nil {
			logErrorWithTimestamp(err)
		}

		deleted, failed := DeleteGraLogs(logs)

		logWithTimestamp(fmt.Sprintf("Deleted %v files, failed to delete %v files\n", deleted, failed))
		logWithTimestamp("Sleeping for one hour\n")
		time.Sleep(1 * time.Hour)
	}
}

func FindGraLogs(dir string, ageCutoff time.Duration) ([]string, error) {
	oldestTime := time.Now().Add(-ageCutoff)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var oldGraLogs []string

	for _, file := range files {
		fileName := file.Name()
		if strings.HasPrefix(fileName, "GRA_") &&
			strings.HasSuffix(fileName, ".log") &&
			file.ModTime().Before(oldestTime) {
			oldGraLogs = append(oldGraLogs, fileName)
		}
	}

	return oldGraLogs, nil
}

func DeleteGraLogs(files []string) (int, int) {
	succeeded := 0
	failed := 0

	for _, file := range files {
		err := os.Remove(file)

		if err == nil {
			succeeded++
		} else {
			logErrorWithTimestamp(err)
			failed++
		}
	}

	return succeeded, failed
}

func logErrorWithTimestamp(err error) {
	fmt.Fprintf(os.Stderr, "[%s] - ", time.Now().Local())
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
}

func logWithTimestamp(format string, args ...interface{}) {
	fmt.Printf("[%s] - ", time.Now().Local())
	if nil == args {
		fmt.Printf(format)
	} else {
		fmt.Printf(format, args...)
	}
}
