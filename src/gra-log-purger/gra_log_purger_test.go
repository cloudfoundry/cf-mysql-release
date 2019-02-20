package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	. "gra-log-purger"
)

var _ = Describe("gra-log-purger", func() {
	var (
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "gra-log-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Remove(tempDir)
	})

	It("requires a graLogDir option", func() {
		cmd := exec.Command(
			graLogPurgerBinPath,
		)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(BeZero())
		Expect(session.Err).To(gbytes.Say(`No gra log directory supplied`))
	})

	It("requires a pidfile option", func() {
		cmd := exec.Command(
			graLogPurgerBinPath,
			"-graLogDir=some/path/to/datadir",
		)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(BeZero())
		Expect(session.Err).To(gbytes.Say(`No pidfile supplied`))
	})

	It("validates graLogDaysToKeep is not less than 0", func() {
		cmd := exec.Command(
			graLogPurgerBinPath,
			"-pidfile="+tempDir+"/gra-log-purger.pid",
			"-graLogDir="+tempDir,
			"-graLogDaysToKeep=-1",
		)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(BeZero())
		Expect(session.Err).To(gbytes.Say(`graLogDaysToKeep should be >= 0`))
	})

	It("manages pid-files", func() {
		cmd := exec.Command(
			graLogPurgerBinPath,
			"-graLogDir="+tempDir,
			"-pidfile="+tempDir+"/gra-log-purger.pid",
		)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		By("Creating a pid-file in the specified location", func() {
			Eventually(func() string {
				return tempDir + "/gra-log-purger.pid"
			}, "1m").Should(BeARegularFile())
		})

		By("Removing the pid-file when terminated cleanly", func() {
			session.Terminate()

			Eventually(session).Should(gexec.Exit(0))

			Expect(tempDir + "/gra-log-purger.pid").NotTo(BeAnExistingFile())
		})
	})

	Context("when GRA log files exist in a directory", func() {
		var (
			expectedRetainedFiles []string
		)
		BeforeEach(func() {
			expectedRetainedFiles = setupGraLogFiles(tempDir, 2048)
		})

		It("removes only the GRA logs", func() {
			cmd := exec.Command(
				graLogPurgerBinPath,
				"-graLogDir="+tempDir,
				"-graLogDaysToKeep=1",
				"-pidfile=/tmp/gra-log-purger.pid",
			)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() ([]string, error) {
				f, err := os.Open(tempDir)
				if err != nil {
					return nil, err
				}

				names, err := f.Readdirnames(-1)
				if err != nil {
					return nil, err
				}

				for i, name := range names {
					names[i] = filepath.Join(tempDir, name)
				}

				return names, nil
			}, "10s").Should(ConsistOf(expectedRetainedFiles))

			session.Terminate()
			Eventually(session).Should(gexec.Exit(0))

			Eventually(func() ([]string, error) {
				f, err := os.Open(tempDir)
				if err != nil {
					return nil, err
				}

				names, err := f.Readdirnames(-1)
				if err != nil {
					return nil, err
				}

				for i, name := range names {
					names[i] = filepath.Join(tempDir, name)
				}

				return names, nil
			}, "1s").Should(ConsistOf(expectedRetainedFiles))
		})
	})
})

var _ = Describe("PurgeGraLogs", func() {
	var (
		err     error
		tempDir string
	)

	BeforeEach(func() {
		tempDir, err = ioutil.TempDir("", "gra-log-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Remove(tempDir)
	})

	Context("When an empty directory is passed in", func() {
		It("returns an empty list and succeeds", func() {
			succeeded, failed, err := PurgeGraLogs(tempDir, time.Hour*24*2)

			Expect(err).NotTo(HaveOccurred())
			Expect(succeeded).To(BeZero())
			Expect(failed).To(BeZero())
		})
	})

	Context("when a directory matches a GRAlog file name", func() {
		BeforeEach(func() {
			Expect(os.Mkdir(filepath.Join(tempDir, "GRA_something.log"), 0755)).To(Succeed())
		})

		It("does not remove directories", func() {
			_, _, err := PurgeGraLogs(tempDir, 0)

			Expect(err).NotTo(HaveOccurred())
			Expect(filepath.Join(tempDir, "GRA_something.log")).Should(BeADirectory())
		})

	})

	Context("When a file that is not a directory is passed in", func() {
		var (
			tempFileFd *os.File
			tempFile   string
		)

		BeforeEach(func() {
			tempFileFd, err = ioutil.TempFile(tempDir, "not-a-directory")
			Expect(err).NotTo(HaveOccurred())
			tempFile = tempFileFd.Name()
		})

		AfterEach(func() {
			os.Remove(tempFile)
		})

		It("returns an out of sandbox error", func() {
			_, _, err := PurgeGraLogs(tempFile, time.Hour*24*2)
			Expect(err).To(BeAssignableToTypeOf(&os.SyscallError{}))
		})
	})
})

func setupGraLogFiles(path string, numberOfGraLogs int) (pathsToRetain []string) {
	// mysql datadir fixtures
	mysqlFixtures := []string{
		filepath.Join(path, "ibdata1"),
		filepath.Join(path, "ib_logfile0"),
		filepath.Join(path, "ib_logfile1"),
		filepath.Join(path, "mysql-bin.000001"),
		filepath.Join(path, "mysql-bin.index"),
	}

	for _, name := range mysqlFixtures {
		Expect(ioutil.WriteFile(name, nil, 0640)).
			To(Succeed())

		fiveDaysAgo := time.Now().Add(-time.Hour * 24 * 5)
		Expect(os.Chtimes(name, fiveDaysAgo, fiveDaysAgo)).
			To(Succeed())
		Expect(os.Chtimes(name, fiveDaysAgo, fiveDaysAgo)).
			To(Succeed())
	}

	pathsToRetain = append(pathsToRetain, mysqlFixtures...)

	// old GRA logs
	for i := 0; i < numberOfGraLogs; i++ {
		graLogPath := filepath.Join(path, fmt.Sprintf("GRA_OLD_%d.log", i))

		Expect(ioutil.WriteFile(graLogPath, nil, 0640)).
			To(Succeed())

		fiveDaysAgo := time.Now().Add(-time.Hour * 24 * 5)
		Expect(os.Chtimes(graLogPath, fiveDaysAgo, fiveDaysAgo)).
			To(Succeed())
	}

	// new GRA logs
	for i := 0; i < 100; i++ {
		graLogPath := filepath.Join(path, fmt.Sprintf("GRA_NEW_%d.log", i))

		Expect(ioutil.WriteFile(graLogPath, nil, 0640)).
			To(Succeed())

		oneHourAgo := time.Now().Add(-time.Hour)
		Expect(os.Chtimes(graLogPath, oneHourAgo, oneHourAgo)).
			To(Succeed())

		pathsToRetain = append(pathsToRetain, graLogPath)
	}

	// non GRA logs
	for i := 0; i < 100; i++ {

		graLogPath := filepath.Join(path, fmt.Sprintf("NOT_A_GRA_%d.log", i))

		Expect(ioutil.WriteFile(graLogPath, nil, 0640)).
			To(Succeed())

		fiveDaysAgo := time.Now().Add(-time.Hour * 24 * 5)
		Expect(os.Chtimes(graLogPath, fiveDaysAgo, fiveDaysAgo)).
			To(Succeed())

		pathsToRetain = append(pathsToRetain, graLogPath)
	}

	return pathsToRetain
}
