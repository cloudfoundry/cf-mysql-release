package database

import (
	"fmt"

	"database/sql"

	"code.cloudfoundry.org/lager"
)

// Sprintf thinks the % sign is a formatting directive, so we escape it with a second %.
const revokeQuery = `REVOKE INSERT, UPDATE, CREATE ON %s.* FROM '%s'@'%%'`

const grantQuery = `GRANT INSERT, UPDATE, CREATE ON %s.* TO '%s'@'%%'`

type Database interface {
	Name() string
	GrantPrivileges() error
	RevokePrivileges() error
	KillActiveConnections() error
}

type database struct {
	name   string
	user   string
	db     *sql.DB
	logger lager.Logger
}

func New(name, user string, db *sql.DB, logger lager.Logger) Database {
	return &database{
		name:   name,
		user:   user,
		db:     db,
		logger: logger,
	}
}

func (d database) Name() string {
	return d.name
}

func (d database) RevokePrivileges() error {
	d.logger.Info(fmt.Sprintf("Revoking privileges to db '%s', user '%s'", d.name, d.user))
	result, err := d.db.Exec(fmt.Sprintf(revokeQuery, d.name, d.user))
	if err != nil {
		return fmt.Errorf("Updating db '%s', user '%s' to revoke privileges: %s", d.name, d.user, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Updating db '%s', user '%s' to revoke privileges: Getting rows affected: %s", d.name, d.user, err.Error())
	}

	d.logger.Info(fmt.Sprintf("Updating db '%s', user '%s' to revoke privileges: Rows affected: %d", d.name, d.user, rowsAffected))

	_, err = d.db.Exec("FLUSH PRIVILEGES")
	if err != nil {
		return fmt.Errorf("Flushing privileges: %s", err.Error())
	}

	return nil
}

func (d database) GrantPrivileges() error {
	d.logger.Info(fmt.Sprintf("Granting privileges to db '%s', user '%s'", d.name, d.user))
	result, err := d.db.Exec(fmt.Sprintf(grantQuery, d.name, d.user))
	if err != nil {
		return fmt.Errorf("Updating db '%s', user '%s' to grant privileges: %s", d.name, d.user, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Updating db '%s', user '%s' to grant privileges: Getting rows affected: %s", d.name, d.user, err.Error())
	}

	d.logger.Info(fmt.Sprintf("Updating db '%s', user '%s' to grant privileges: Rows affected: %d", d.name, d.user, rowsAffected))

	_, err = d.db.Exec("FLUSH PRIVILEGES")
	if err != nil {
		return fmt.Errorf("Flushing privileges: %s", err.Error())
	}

	return nil
}

// ResetActivePrivileges flushes the privileges and kills all active connections to this database.
// New connections will get the new privileges.
func (d database) KillActiveConnections() error {
	d.logger.Info(fmt.Sprintf("Killing active connections to database '%s'", d.name))

	rows, err := d.db.Query("SELECT ID FROM INFORMATION_SCHEMA.PROCESSLIST WHERE DB = ?", d.name)
	if err != nil {
		return fmt.Errorf("Getting list of open connections to database '%s': %s", d.name, err.Error())
	}
	//TODO: untested Close, due to limitation of sqlmock: https://github.com/DATA-DOG/go-sqlmock/issues/15
	defer rows.Close()
	for rows.Next() {
		var connectionID int64
		if err := rows.Scan(&connectionID); err != nil {
			//TODO: untested error case, due to limitation of sqlmock: https://github.com/DATA-DOG/go-sqlmock/issues/13
			return fmt.Errorf("Scanning open connections to database '%s': %s", d.name, err.Error())
		}

		d.logger.Debug(fmt.Sprintf("Killing active connection %d to database '%s'", connectionID, d.name))
		_, err := d.db.Exec("KILL CONNECTION ?", connectionID)
		if err != nil {
			d.logger.Error(fmt.Sprintf("Failed to kill active connection %d to database '%s'", connectionID, d.name), err)
		}
	}
	//TODO: untested error case, due to limitation of sqlmock: https://github.com/DATA-DOG/go-sqlmock/issues/13
	if err := rows.Err(); err != nil {
		return fmt.Errorf("Reading open connections to database '%s': %s", d.name, err.Error())
	}

	return nil
}
