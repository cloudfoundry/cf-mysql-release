package cf

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
	. "github.com/onsi/gomega/gexec"
)

var Cf = func(args ...string) *Session {
	println(fmt.Sprintf("Calling cf with args %v", args))
	return runner.Run("cf", args...)
}
