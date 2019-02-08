package main_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "gra-log-purger"
)

var _ = Describe("FindGraLogs", func() {
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
			files, err := FindGraLogs(tempDir, time.Hour*24*2)

			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(BeEmpty())
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
			_, err := FindGraLogs(tempFile, time.Hour*24*2)
			Expect(err).To(BeAssignableToTypeOf(&os.SyscallError{}))
		})
	})

	Context("When a directory with GRA files is passed in", func() {
		BeforeEach(func() {
			// old GRA logs
			for i := 0; i < 100; i++ {
				tempGraFileFd, err := ioutil.TempFile(tempDir, "GRA_OLD_*.log")
				Expect(err).NotTo(HaveOccurred())

				tempGraFile := tempGraFileFd.Name()
				Expect(tempGraFileFd.Close()).To(Succeed())

				fiveDaysAgo := time.Now().Add(-time.Hour * 24 * 5)
				err = os.Chtimes(tempGraFile, fiveDaysAgo, fiveDaysAgo)
				Expect(err).NotTo(HaveOccurred())
			}

			// new GRA logs
			for i := 0; i < 100; i++ {
				tempGraFileFd, err := ioutil.TempFile(tempDir, "GRA_NEW_*.log")
				Expect(err).NotTo(HaveOccurred())

				tempGraFile := tempGraFileFd.Name()
				Expect(tempGraFileFd.Close()).To(Succeed())

				oneDayAgo := time.Now().Add(-time.Hour * 24)
				err = os.Chtimes(tempGraFile, oneDayAgo, oneDayAgo)
				Expect(err).NotTo(HaveOccurred())

			}

			// non GRA logs
			for i := 0; i < 100; i++ {
				tempNotGraFileFd, err := ioutil.TempFile(tempDir, "NOT_A_GRA_*.log")
				Expect(err).NotTo(HaveOccurred())

				tempNotGraFile := tempNotGraFileFd.Name()
				Expect(tempNotGraFileFd.Close()).To(Succeed())

				fiveDaysAgo := time.Now().Add(-time.Hour * 24 * 5)
				err = os.Chtimes(tempNotGraFile, fiveDaysAgo, fiveDaysAgo)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("returns GRA log files older than the cutoff", func() {
			graLogs, err := FindGraLogs(tempDir, time.Hour*24*2)
			Expect(err).NotTo(HaveOccurred())
			Expect(graLogs).To(HaveLen(100))

			for _, log := range graLogs {
				// only returns the OLD files, no NEW and no NOT_A_GRA_LOG
				Expect(log).To(ContainSubstring("GRA_OLD"))
			}
		})

		It("doesn't return nested files", func() {
			tempSubDir, err := ioutil.TempDir(tempDir, "sub-directory")
			Expect(err).NotTo(HaveOccurred())

			// create an OLD GRA log inside the sub directory
			tempGraFileFd, err := ioutil.TempFile(tempSubDir, "GRA_OLD_SUB_DIR_*.log")
			Expect(err).NotTo(HaveOccurred())

			tempGraFile := tempGraFileFd.Name()
			Expect(tempGraFileFd.Close()).To(Succeed())

			fiveDaysAgo := time.Now().Add(-time.Hour * 24 * 5)
			err = os.Chtimes(tempGraFile, fiveDaysAgo, fiveDaysAgo)
			Expect(err).NotTo(HaveOccurred())

			// check that we don't find this file
			graLogs, err := FindGraLogs(tempDir, time.Hour*24*2)
			Expect(err).NotTo(HaveOccurred())
			Expect(graLogs).To(HaveLen(100))

			for _, log := range graLogs {
				// only returns the OLD files, no NEW and no NOT_A_GRA_LOG
				Expect(log).NotTo(ContainSubstring("GRA_OLD_SUB_DIR"))
			}
		})

	})

})

var _ = Describe("DeleteGraLogs", func() {
	var (
		err     error
		tempDir string
		fileA   *os.File
		fileB   *os.File
	)

	BeforeEach(func() {
		tempDir, err = ioutil.TempDir("", "gra-log-test")
		Expect(err).NotTo(HaveOccurred())

		fileA, err = ioutil.TempFile(tempDir, "A")
		Expect(err).NotTo(HaveOccurred())

		fileB, err = ioutil.TempFile(tempDir, "B")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Remove(tempDir)
	})

	Context("When an empty list is passed in", func() {
		It("does nothing and succeeds", func() {
			deleted, failed := DeleteGraLogs(nil)
			Expect(deleted).To(BeZero())
			Expect(failed).To(BeZero())
		})
	})

	It("deletes all the file paths passed in", func() {
		deleted, failed := DeleteGraLogs([]string{
			fileA.Name(),
			fileB.Name(),
		})

		Expect(deleted).To(Equal(2))
		Expect(failed).To(BeZero())
	})

	Context("When a non-existent path is passed in", func() {
		It("deletes the valid paths and counts the failure", func() {
			deleted, failed := DeleteGraLogs([]string{
				fileA.Name(),
				"FAKE-FILE",
				fileB.Name(),
			})

			Expect(deleted).To(Equal(2))
			Expect(failed).To(Equal(1))
		})
	})
})
