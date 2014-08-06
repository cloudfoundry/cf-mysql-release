package galera_helper_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGalera_helper(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Galera Helper Suite")
}
