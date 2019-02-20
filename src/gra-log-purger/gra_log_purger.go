package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		graLogDir        string
		graLogDaysToKeep int
		pidfile          string
	)

	flag.StringVar(&graLogDir,
		"graLogDir",
		"",
		"Specifies the directory from which to purge GRA log files.",
	)

	flag.IntVar(&graLogDaysToKeep,
		"graLogDaysToKeep",
		60,
		"Specifies the maximum age of the GRA log files allowed.",
	)

	flag.StringVar(&pidfile,
		"pidfile",
		"",
		"The location for the pidfile",
	)

	flag.Parse()

	if graLogDir == "" {
		logErrorWithTimestamp(fmt.Errorf("No gra log directory supplied"))
		os.Exit(1)
	}

	if pidfile == "" {
		logErrorWithTimestamp(fmt.Errorf("No pidfile supplied"))
		os.Exit(1)
	}

	if graLogDaysToKeep < 0 {
		logErrorWithTimestamp(fmt.Errorf("graLogDaysToKeep should be >= 0"))
		os.Exit(1)
	}

	err := ioutil.WriteFile(pidfile, []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		panic(err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	logWithTimestamp("Will purge old GRA logs once every hour\n")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	cleanup := func() {
		ageCutoff := time.Duration(graLogDaysToKeep*24) * time.Hour
		deleted, failed, err := PurgeGraLogs(graLogDir, ageCutoff)
		if err != nil {
			logErrorWithTimestamp(err)
		} else {
			logWithTimestamp(fmt.Sprintf("Deleted %v files, failed to delete %v files\n", deleted, failed))
		}

		logWithTimestamp("Sleeping for one hour\n")
	}

	cleanup()

	for {
		select {
		case sig := <-sigCh:
			logWithTimestamp("%s", sig)
			os.Remove(pidfile)
			os.Exit(0)
		case <-ticker.C:
			cleanup()
		}
	}
}

func isOldGraLog(file os.FileInfo, oldestTime time.Time) bool {
	if file.IsDir() == false &&
		strings.HasPrefix(file.Name(), "GRA_") &&
		strings.HasSuffix(file.Name(), ".log") &&
		file.ModTime().Before(oldestTime) {
		return true
	}

	return false
}

func PurgeGraLogs(dir string, ageCutoff time.Duration) (int, int, error) {
	succeeded := 0
	failed := 0

	handle, err := os.Open(dir)
	if err != nil {
		return succeeded, failed, err
	}

	oldestTime := time.Now().Add(-ageCutoff)
	for {
		files, err := handle.Readdir(1024)
		if err == io.EOF {
			break
		} else if err != nil {
			return succeeded, failed, err
		}

		for _, file := range files {
			fileName := file.Name()
			if isOldGraLog(file, oldestTime) {
				if err := os.Remove(filepath.Join(dir, fileName)); err != nil {
					logErrorWithTimestamp(err)
					failed++
				} else {
					succeeded++
				}
			}
		}
	}

	return succeeded, failed, nil
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
