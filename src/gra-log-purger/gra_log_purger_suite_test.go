package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUpgrader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gra Log Purger Suite")
}
