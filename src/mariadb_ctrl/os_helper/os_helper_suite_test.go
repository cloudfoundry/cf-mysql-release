package os_helper_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOs_helper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Os_helper Suite")
}
