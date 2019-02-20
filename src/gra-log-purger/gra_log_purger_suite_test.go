package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestUpgrader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gra Log Purger Suite")
}

var (
	graLogPurgerBinPath string
)

var _ = BeforeSuite(func() {
	var err error
	graLogPurgerBinPath, err = gexec.Build("gra-log-purger")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.Kill()
	gexec.CleanupBuildArtifacts()
})