package enforcer_test

import (
	"code.cloudfoundry.org/lager/lagertest"
	"quota-enforcer/database"
	"quota-enforcer/database/databasefakes"
	. "quota-enforcer/enforcer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Enforcer", func() {
	var (
		enforcer         Enforcer
		fakeViolatorRepo *databasefakes.FakeRepo
		fakeReformerRepo *databasefakes.FakeRepo
		logger           *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("Enforcer test")
		fakeViolatorRepo = &databasefakes.FakeRepo{}
		fakeReformerRepo = &databasefakes.FakeRepo{}
		enforcer = NewEnforcer(fakeViolatorRepo, fakeReformerRepo, logger)
	})

	Context("when there are no violators", func() {
		It("does not revoke privileges for anyone", func() {
			err := enforcer.EnforceOnce()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeViolatorRepo.AllCallCount()).To(Equal(1))
		})
	})

	Context("when there are violators", func() {
		var fakeViolators []database.Database

		BeforeEach(func() {
			fakeViolators = []database.Database{
				&databasefakes.FakeDatabase{},
				&databasefakes.FakeDatabase{},
			}
			fakeViolatorRepo.AllReturns(fakeViolators, nil)
		})

		It("revokes privileges on the violators", func() {
			err := enforcer.EnforceOnce()
			Expect(err).NotTo(HaveOccurred())

			for _, db := range fakeViolators {
				fakeDB := db.(*databasefakes.FakeDatabase)
				Expect(fakeDB.RevokePrivilegesCallCount()).To(Equal(1))
				Expect(fakeDB.KillActiveConnectionsCallCount()).To(Equal(1))
			}
		})
	})

	Context("when there are no reformers", func() {
		It("does not grant privileges for anyone", func() {
			err := enforcer.EnforceOnce()
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReformerRepo.AllCallCount()).To(Equal(1))
		})
	})

	Context("when there are reformers", func() {
		var fakeReformers []database.Database

		BeforeEach(func() {
			fakeReformers = []database.Database{
				&databasefakes.FakeDatabase{},
				&databasefakes.FakeDatabase{},
			}
			fakeReformerRepo.AllReturns(fakeReformers, nil)
		})

		It("grants privileges on the reformers", func() {
			err := enforcer.EnforceOnce()
			Expect(err).NotTo(HaveOccurred())

			for _, db := range fakeReformers {
				fakeDB := db.(*databasefakes.FakeDatabase)
				Expect(fakeDB.GrantPrivilegesCallCount()).To(Equal(1))
				Expect(fakeDB.KillActiveConnectionsCallCount()).To(Equal(1))
			}
		})
	})
})
