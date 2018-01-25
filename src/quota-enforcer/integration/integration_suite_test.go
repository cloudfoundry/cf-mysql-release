package enforcer_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"quota-enforcer/config"
	"quota-enforcer/database"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pivotal-cf-experimental/service-config"
)

var brokerDBName string
var c config.Config
var binaryPath string

var tempDir string
var configPath string

var adminDB *sql.DB
var adminCreds AdminCredentials
var initConfig config.Config

type AdminCredentials struct {
	User     string `yaml:"AdminUser" validate:"nonzero"`
	Password string `yaml:"AdminPassword" validate:"nonzero"`
}

func TestEnforcer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Enforcer Suite")
}

func newConfig(dbName string) config.Config {
	serviceConfig := service_config.New()

	var cfg config.Config
	err := serviceConfig.Read(&cfg)
	Expect(err).ToNot(HaveOccurred())

	cfg.DBName = dbName
	cfg.IgnoredUsers = append(cfg.IgnoredUsers, "fake-admin-user")
	cfg.PauseInSeconds = 1

	return cfg
}

func adminCredentials() AdminCredentials {
	serviceConfig := service_config.New()

	var adminCreds AdminCredentials
	err := serviceConfig.Read(&adminCreds)
	Expect(err).ToNot(HaveOccurred())

	return adminCreds
}

var _ = BeforeSuite(func() {
	initConfig = newConfig("")

	brokerDBName = uuidWithUnderscores("db")
	c = newConfig(brokerDBName)

	adminCreds = adminCredentials()

	adminDB, err := database.NewConnection(adminCreds.User, adminCreds.Password, initConfig.Host, initConfig.Port, initConfig.DBName)
	Expect(err).ToNot(HaveOccurred())
	defer adminDB.Close()

	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.DBName))
	Expect(err).ToNot(HaveOccurred())

	db, err := database.NewConnection(c.User, c.Password, c.Host, c.Port, c.DBName)
	Expect(err).ToNot(HaveOccurred())

	for _, ignoredUser := range initConfig.IgnoredUsers {
		_, err = adminDB.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON *.* TO '%s' IDENTIFIED BY '%s' WITH GRANT OPTION",
			ignoredUser,
			"password",
		))
	}

	_, err = adminDB.Exec("FLUSH PRIVILEGES")
	Expect(err).NotTo(HaveOccurred())

	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS service_instances (
    id int(11) NOT NULL AUTO_INCREMENT,
    guid varchar(255),
    plan_guid varchar(255),
    max_storage_mb int(11) NOT NULL DEFAULT '0',
    db_name varchar(255),
    PRIMARY KEY (id)
	)`)
	Expect(err).ToNot(HaveOccurred())

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS read_only_users (
    id int(11) NOT NULL AUTO_INCREMENT,
    username varchar(255),
    grantee varchar(255),
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL,
    PRIMARY KEY (id)
	)`)
	Expect(err).ToNot(HaveOccurred())

	binaryPath, err = gexec.Build("quota-enforcer", "-race")
	Expect(err).ToNot(HaveOccurred())

	_, err = os.Stat(binaryPath)
	if err != nil {
		Expect(os.IsExist(err)).To(BeTrue())
	}

	tempDir, err = ioutil.TempDir(os.TempDir(), "quota-enforcer-integration-test")
	Expect(err).NotTo(HaveOccurred())

	configPath = filepath.Join(tempDir, "quotaEnforcerConfig.yml")
	writeConfig()
})

var _ = AfterSuite(func() {

	// We don't need to handle an error cleaning up the tempDir
	_ = os.RemoveAll(tempDir)

	gexec.CleanupBuildArtifacts()

	_, err := os.Stat(binaryPath)
	if err != nil {
		Expect(os.IsExist(err)).To(BeFalse())
	}

	db, err := database.NewConnection(c.User, c.Password, c.Host, c.Port, c.DBName)
	Expect(err).ToNot(HaveOccurred())
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", brokerDBName))
	Expect(err).ToNot(HaveOccurred())

	adminDB, err := database.NewConnection(adminCreds.User, adminCreds.Password, initConfig.Host, initConfig.Port, initConfig.DBName)
	Expect(err).ToNot(HaveOccurred())
	defer adminDB.Close()

	for _, ignoredUser := range initConfig.IgnoredUsers {
		_, err = adminDB.Exec(fmt.Sprintf("DROP USER IF EXISTS '%s'",
			ignoredUser,
		))
		Expect(err).ToNot(HaveOccurred())
	}
})

func startEnforcerWithFlags(flags ...string) *gexec.Session {

	flags = append(
		flags,
		fmt.Sprintf("-configPath=%s", configPath),
		"-logLevel=debug",
	)

	command := exec.Command(
		binaryPath,
		flags...,
	)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session
}

func runEnforcerContinuously(flags ...string) *gexec.Session {
	session := startEnforcerWithFlags(flags...)
	Eventually(session.Out).Should(gbytes.Say("Running continuously"))
	return session
}

func runEnforcerOnce() {
	session := startEnforcerWithFlags("-runOnce")

	Eventually(session.Out).Should(gbytes.Say("Running once"))
	// Wait for the process to finish naturally.
	// This should not take a long time
	session.Wait(5 * time.Second)
	Expect(session.ExitCode()).To(Equal(0), string(session.Err.Contents()))
}

func writeConfig() {
	fileToWrite, err := os.Create(configPath)
	Expect(err).ToNot(HaveOccurred())

	bytes, err := json.MarshalIndent(c, "", "  ")
	Expect(err).ToNot(HaveOccurred())

	_, err = fileToWrite.Write(bytes)
	Expect(err).ToNot(HaveOccurred())
}

func uuidWithUnderscores(prefix string) string {
	id, err := uuid.NewV4()
	Expect(err).ToNot(HaveOccurred())
	idString := fmt.Sprintf("%s_%s", prefix, id.String())
	return strings.Replace(idString, "-", "_", -1)
}
