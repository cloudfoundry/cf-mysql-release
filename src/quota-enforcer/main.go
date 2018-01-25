package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tedsuo/ifrit"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf-experimental/service-config"
	"quota-enforcer/clock"
	"quota-enforcer/config"
	"quota-enforcer/database"
	"quota-enforcer/enforcer"
)

func main() {
	serviceConfig := service_config.New()

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	runOnce := flags.Bool("runOnce", false, "Run only once instead of continuously")
	pidFile := flags.String("pidFile", "", "Location of pid file")
	serviceConfig.AddFlags(flags)
	cflager.AddFlags(flags)
	flags.Parse(os.Args[1:])
	logger, _ := cflager.New("Quota Enforcer")

	var config config.Config
	err := serviceConfig.Read(&config)
	if err != nil {
		logger.Fatal("Failed to read config", err)
	}

	err = config.Validate()
	if err != nil {
		logger.Fatal("Invalid config", err)
	}

	adminUser := config.User
	brokerDBName := config.DBName

	db, err := database.NewConnection(adminUser, config.Password, config.Host, config.Port, brokerDBName)
	if db != nil {
		defer db.Close()
	}

	if err != nil {
		logger.Fatal("Failed to open database connection", err)
	}

	logger.Info(
		"Database connection established.",
		lager.Data{
			"Host":         config.Host,
			"Port":         config.Port,
			"User":         adminUser,
			"DatabaseName": brokerDBName,
		})
	ignoredUsers := []string{adminUser}
	ignoredUsers = append(ignoredUsers, config.IgnoredUsers...)

	violatorRepo := database.NewViolatorRepo(brokerDBName, ignoredUsers, db, logger)
	reformerRepo := database.NewReformerRepo(brokerDBName, ignoredUsers, db, logger)

	e := enforcer.NewEnforcer(violatorRepo, reformerRepo, logger)
	r := enforcer.NewRunner(
		e,
		clock.DefaultClock(),
		time.Duration(config.PauseInSeconds)*time.Second,
		logger,
	)

	if *runOnce {
		logger.Info("Running once")

		err := e.EnforceOnce()
		if err != nil {
			logger.Info(fmt.Sprintf("Quota Enforcing Failed: %s", err.Error()))
		}
	} else {
		process := ifrit.Invoke(r)
		logger.Info("Running continuously")

		// Write pid file once we are running continuously
		if *pidFile != "" {
			pid := os.Getpid()
			err = writePidFile(pid, *pidFile)
			if err != nil {
				logger.Fatal("Cannot write pid to file", err, lager.Data{"pidFile": pidFile, "pid": pid})
			}
			logger.Info("Wrote pid to file", lager.Data{"pidFile": pidFile, "pid": pid})
		}

		err := <-process.Wait()
		if err != nil {
			logger.Fatal("Quota Enforcing Failed", err)
		}
	}
}

func writePidFile(pid int, pidFile string) error {
	return ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}
