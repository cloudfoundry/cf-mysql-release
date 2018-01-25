package database

import (
	"fmt"

	"database/sql"

	"code.cloudfoundry.org/lager"
)

type Repo interface {
	All() ([]Database, error)
}

type repo struct {
	query      string
	parameters []string
	db         *sql.DB
	logger     lager.Logger
	logTag     string
}

func newRepo(query string, parameters []string, db *sql.DB, logger lager.Logger, logTag string) Repo {
	return &repo{
		query:      query,
		parameters: parameters,
		db:         db,
		logger:     logger,
		logTag:     logTag,
	}
}

func (r repo) All() ([]Database, error) {
	r.logger.Debug(fmt.Sprintf("Executing '%s'.All", r.logTag))

	databases := []Database{}

	parametersInterface := make([]interface{}, len(r.parameters))
	for i, v := range r.parameters {
		parametersInterface[i] = v
	}

	rows, err := r.db.Query(r.query, parametersInterface...)
	if err != nil {
		return databases, fmt.Errorf("Error executing '%s'.All: %s", r.logTag, err.Error())
	}

	r.logger.Debug(fmt.Sprintf("Executing '%s'.All - completed", r.logTag))

	//TODO: untested Close, due to limitation of sqlmock: https://github.com/DATA-DOG/go-sqlmock/issues/15
	defer rows.Close()

	for rows.Next() {
		var dbName, dbUser string
		if err := rows.Scan(&dbName, &dbUser); err != nil {
			//TODO: untested error case, due to limitation of sqlmock: https://github.com/DATA-DOG/go-sqlmock/issues/13
			return databases, fmt.Errorf("Scanning result row of '%s'.All: %s", r.logTag, err.Error())
		}

		databases = append(databases, New(dbName, dbUser, r.db, r.logger))
	}
	//TODO: untested error case, due to limitation of sqlmock: https://github.com/DATA-DOG/go-sqlmock/issues/13
	if err := rows.Err(); err != nil {
		return databases, fmt.Errorf("Reading result row of '%s'.All: %s", r.logTag, err.Error())
	}

	r.logger.Debug(
		"returned databases",
		lager.Data{
			"databases": databases,
			"logTag":    r.logTag,
		},
	)

	return databases, nil
}
