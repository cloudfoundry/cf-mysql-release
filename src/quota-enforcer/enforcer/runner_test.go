package enforcer_test

import (
	"errors"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
	"quota-enforcer/clock/clockfakes"
	enforcerPkg "quota-enforcer/enforcer"
	"quota-enforcer/enforcer/enforcerfakes"
)

var _ = Describe("Runner", func() {

	var (
		enforcer *enforcerfakes.FakeEnforcer
		clock    *clockfakes.FakeClock
		logger   *lagertest.TestLogger
		pause    time.Duration
		runner   ifrit.Runner
		signals  chan os.Signal
		ready    chan struct{}
	)

	BeforeEach(func() {
		enforcer = &enforcerfakes.FakeEnforcer{}
		clock = &clockfakes.FakeClock{}
		pause = 1 * time.Second
		logger = lagertest.NewTestLogger("Runner test")
		runner = enforcerPkg.NewRunner(enforcer, clock, pause, logger)

		signals = make(chan os.Signal, 1)
		ready = make(chan struct{})
		go func() {
			<-ready
			signals <- os.Interrupt
		}()

		clock.AfterStub = func(d time.Duration) <-chan time.Time {
			return time.After(1 * time.Millisecond)
		}
	})

	It("runs the enforcer", func() {
		runner.Run(signals, ready)
		Expect(enforcer.EnforceOnceCallCount()).To(BeNumerically(">", 0))
	})

	It("sleeps between calls to the enforcer", func() {
		runner.Run(signals, ready)
		Expect(clock.AfterCallCount()).To(BeNumerically(">", 0))
		for _, sleep := range clock.Invocations()["After"] {
			Expect(sleep[0]).To(Equal(pause))
		}
	})

	Context("when the enforcer errors", func() {
		It("logs the error", func() {
			enforcer.EnforceOnceStub = func() error {
				return errors.New("failed to uphold the law")
			}
			runner.Run(signals, ready)
			Expect(logger.TestSink.LogMessages()).To(
				ContainElement(ContainSubstring("Enforcing Failed")))

			for _, entry := range logger.TestSink.Logs() {
				if strings.Contains(
					entry.Message, "Enforcing Failed") &&
					entry.LogLevel != lager.ERROR {
					Fail("error logged with non-error log level")
				}
			}
		})
	})

})
