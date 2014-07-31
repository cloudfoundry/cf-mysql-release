package mariadb_start_manager_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMariadb_start_manager(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "MariaDB Start Manager Suite")
}
