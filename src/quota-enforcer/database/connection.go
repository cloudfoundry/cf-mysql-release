package database

import (
	"database/sql"
	"fmt"
)

func NewConnection(username, password, host string, port int, dbName string) (*sql.DB, error) {

	var userPass string
	if password != "" {
		userPass = fmt.Sprintf("%s:%s", username, password)
	} else {
		userPass = username
	}

	return sql.Open("mysql", fmt.Sprintf(
		"%s@tcp(%s:%d)/%s",
		userPass,
		host,
		port,
		dbName,
	))
}
