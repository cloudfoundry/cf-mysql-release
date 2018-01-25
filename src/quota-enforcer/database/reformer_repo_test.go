package database_test

import (
	. "quota-enforcer/database"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"

	"errors"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/DATA-DOG/go-sqlmock"
)

var _ = Describe("ReformerRepo", func() {

	const (
		brokerDBName = "fake_broker_db_name"
		adminUser    = "fake_admin_user"
		readOnlyUser = "fake_read_only_user"
	)

	var (
		logger *lagertest.TestLogger
		repo   Repo
		fakeDB *sql.DB
		mock   sqlmock.Sqlmock
	)

	BeforeEach(func() {
		var err error
		fakeDB, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())

		logger = lagertest.NewTestLogger("ReformerRepo test")
		ignoredUsers := []string{adminUser, readOnlyUser}
		repo = NewReformerRepo(brokerDBName, ignoredUsers, fakeDB, logger)
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Describe("All", func() {
		var (
			tableSchemaColumns = []string{"db", "user"}
			matchAny           = ".*"
		)

		It("returns a list of databases that have come under their quota", func() {
			mock.ExpectQuery(matchAny).
				WithArgs().
				WillReturnRows(
					sqlmock.NewRows(tableSchemaColumns).
						AddRow("fake-database-1", "cf_fake-user-1").
						AddRow("fake-database-2", "cf_fake-user-2"))

			reformers, err := repo.All()
			Expect(err).ToNot(HaveOccurred())

			Expect(reformers).To(ConsistOf(
				New("fake-database-1", "cf_fake-user-1", fakeDB, logger),
				New("fake-database-2", "cf_fake-user-2", fakeDB, logger),
			))
		})

		It("passes ignored users as ordered parameters", func() {
			mock.ExpectQuery("NOT IN \\(\\?,\\?\\)").
				WithArgs().
				WillReturnRows(sqlmock.NewRows(tableSchemaColumns))

			_, err := repo.All()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when there are no reformers", func() {
			BeforeEach(func() {
				mock.ExpectQuery(matchAny).
					WithArgs().
					WillReturnRows(sqlmock.NewRows(tableSchemaColumns))
			})

			It("returns an empty list", func() {
				violators, err := repo.All()
				Expect(err).ToNot(HaveOccurred())

				Expect(violators).To(BeEmpty())
			})
		})

		Context("when the db query fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(matchAny).
					WithArgs().
					WillReturnError(errors.New("fake-query-error"))
			})

			It("returns an error", func() {
				_, err := repo.All()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-query-error"))
			})
		})
	})
})
