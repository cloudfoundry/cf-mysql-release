package database_test

import (
	. "quota-enforcer/database"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"database/sql"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/DATA-DOG/go-sqlmock"
)

var _ = Describe("Database", func() {

	const (
		dbName = "fake-db-name"
		dbUser = "fake-db-user"
	)

	var (
		logger                 *lagertest.TestLogger
		database               Database
		fakeDB                 *sql.DB
		flushPrivilegesPattern = "FLUSH PRIVILEGES"
		mock                   sqlmock.Sqlmock
	)

	BeforeEach(func() {
		var err error
		fakeDB, mock, err = sqlmock.New()
		Expect(err).ToNot(HaveOccurred())

		logger = lagertest.NewTestLogger("Database test")
		database = New(dbName, dbUser, fakeDB, logger)
	})

	AfterEach(func() {
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	Describe("RevokePrivileges", func() {
		var (
			revokePrivilegesPattern = `REVOKE INSERT, UPDATE, CREATE ON fake-db-name.\* FROM 'fake-db-user'@'\%'`
		)

		It("makes a sql query to revoke privileges on a database and then flushes privileges", func() {
			mock.ExpectExec(revokePrivilegesPattern).
				WillReturnResult(sqlmock.NewResult(-1, 1))

			mock.ExpectExec(flushPrivilegesPattern).
				WithArgs().
				WillReturnResult(sqlmock.NewResult(-1, 1))

			err := database.RevokePrivileges()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the query fails", func() {
			BeforeEach(func() {
				mock.ExpectExec(revokePrivilegesPattern).
					WillReturnError(errors.New("fake-query-error"))
			})

			It("returns an error", func() {
				err := database.RevokePrivileges()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-query-error"))
				Expect(err.Error()).To(ContainSubstring(dbName))
				Expect(err.Error()).To(ContainSubstring(dbUser))
			})
		})

		Context("when getting the number of affected rows fails", func() {
			BeforeEach(func() {
				mock.ExpectExec(revokePrivilegesPattern).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("fake-rows-affected-error")))
			})

			It("returns an error", func() {
				err := database.RevokePrivileges()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-rows-affected-error"))
				Expect(err.Error()).To(ContainSubstring("Getting rows affected"))
				Expect(err.Error()).To(ContainSubstring(dbName))
				Expect(err.Error()).To(ContainSubstring(dbUser))
			})
		})

		Context("when flushing privileges fails", func() {
			BeforeEach(func() {
				mock.ExpectExec(revokePrivilegesPattern).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				mock.ExpectExec(flushPrivilegesPattern).
					WithArgs().
					WillReturnError(errors.New("fake-flush-error"))
			})

			It("returns an error", func() {
				err := database.RevokePrivileges()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-flush-error"))
			})
		})
	})

	Describe("GrantPrivileges", func() {
		var (
			grantPrivilegesPattern = `GRANT INSERT, UPDATE, CREATE ON fake-db-name.\* TO 'fake-db-user'@'\%'`
		)

		It("grants privileges to the database", func() {
			mock.ExpectExec(grantPrivilegesPattern).
				WillReturnResult(sqlmock.NewResult(-1, 1))

			mock.ExpectExec(flushPrivilegesPattern).
				WillReturnResult(sqlmock.NewResult(-1, 1))

			err := database.GrantPrivileges()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the query fails", func() {
			BeforeEach(func() {
				mock.ExpectExec(grantPrivilegesPattern).
					WillReturnError(errors.New("fake-query-error"))
			})

			It("returns an error", func() {
				err := database.GrantPrivileges()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-query-error"))
				Expect(err.Error()).To(ContainSubstring(dbName))
				Expect(err.Error()).To(ContainSubstring(dbUser))
			})
		})

		Context("when getting the number of affected rows fails", func() {
			BeforeEach(func() {
				mock.ExpectExec(grantPrivilegesPattern).
					WillReturnResult(sqlmock.NewErrorResult(errors.New("fake-rows-affected-error")))
			})

			It("returns an error", func() {
				err := database.GrantPrivileges()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-rows-affected-error"))
				Expect(err.Error()).To(ContainSubstring("Getting rows affected"))
				Expect(err.Error()).To(ContainSubstring(dbName))
				Expect(err.Error()).To(ContainSubstring(dbUser))
			})
		})

		Context("when flushing privileges fails", func() {
			BeforeEach(func() {
				mock.ExpectExec(grantPrivilegesPattern).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				mock.ExpectExec(flushPrivilegesPattern).
					WithArgs().
					WillReturnError(errors.New("fake-flush-error"))
			})

			It("returns an error", func() {
				err := database.GrantPrivileges()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-flush-error"))
			})
		})

	})

	Describe("KillActiveConnections", func() {
		var (
			processListColumns    = []string{"ID"}
			processQueryPattern   = `SELECT ID FROM INFORMATION_SCHEMA.PROCESSLIST WHERE DB = \?$`
			killConnectionPattern = "KILL CONNECTION \\?"
		)

		It("kills all active connections to DB", func() {
			mock.ExpectQuery(processQueryPattern).
				WithArgs(dbName).
				WillReturnRows(sqlmock.NewRows(processListColumns).AddRow(1).AddRow(123))

			mock.ExpectExec(killConnectionPattern).
				WithArgs(1).
				WillReturnResult(sqlmock.NewResult(-1, 1))

			mock.ExpectExec(killConnectionPattern).
				WithArgs(123).
				WillReturnResult(sqlmock.NewResult(-1, 1))

			err := database.KillActiveConnections()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when there are no active connections to the database", func() {
			It("does not kill any connections", func() {
				mock.ExpectQuery(processQueryPattern).
					WithArgs(dbName).
					WillReturnRows(sqlmock.NewRows(processListColumns))

				err := database.KillActiveConnections()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when there is only one active connections to the database", func() {
			It("kills the active connection", func() {
				mock.ExpectQuery(processQueryPattern).
					WithArgs(dbName).
					WillReturnRows(sqlmock.NewRows(processListColumns).AddRow(123))

				mock.ExpectExec(killConnectionPattern).
					WithArgs(123).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				err := database.KillActiveConnections()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when querying for active connections fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(processQueryPattern).
					WithArgs(dbName).
					WillReturnError(errors.New("fake-query-error"))
			})

			It("returns an error", func() {
				err := database.KillActiveConnections()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-query-error"))
				Expect(err.Error()).To(ContainSubstring(dbName))
			})
		})

		Context("when killing a connection fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(processQueryPattern).
					WithArgs(dbName).
					WillReturnRows(sqlmock.NewRows(processListColumns).AddRow(1).AddRow(2).AddRow(3))
			})

			It("kills all other active connections", func() {
				mock.ExpectExec(killConnectionPattern).
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				mock.ExpectExec(killConnectionPattern).
					WithArgs(2).
					WillReturnError(errors.New("fake-exec-error"))

				mock.ExpectExec(killConnectionPattern).
					WithArgs(3).
					WillReturnResult(sqlmock.NewResult(-1, 1))

				err := database.KillActiveConnections()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
