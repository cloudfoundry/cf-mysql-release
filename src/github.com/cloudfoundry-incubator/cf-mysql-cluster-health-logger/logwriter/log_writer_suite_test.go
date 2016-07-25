package logwriter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogWriter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogWriter Suite")
}
