package database

import (
	"database/sql"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
)

const violatorsQueryPattern = `
SELECT violators.name AS violator_db, violators.user AS violator_user
FROM (
	SELECT dbs.name, dbs.user, tables.data_length, tables.index_length
	FROM   (
		SELECT DISTINCT table_schema AS name, replace(substring_index(grantee, '@', 1), "'", '') AS user
		FROM information_schema.schema_privileges
		WHERE privilege_type IN ('INSERT', 'UPDATE', 'CREATE')
		AND replace(substring_index(grantee, '@', 1), "'", '') NOT IN (%s)
	) AS dbs
	JOIN %s.service_instances AS instances ON dbs.name = instances.db_name COLLATE utf8_general_ci
	JOIN information_schema.tables AS tables ON tables.table_schema = dbs.name
	GROUP BY dbs.user
	HAVING ROUND(SUM(COALESCE(tables.data_length + tables.index_length,0) / 1024 / 1024), 1) >= MAX(instances.max_storage_mb)
) AS violators
`

func NewViolatorRepo(brokerDBName string, ignoredUsers []string, db *sql.DB, logger lager.Logger) Repo {
	ignoredUsersPlaceholders := strings.Join(strings.Split(strings.Repeat("?", len(ignoredUsers)), ""), ",")
	query := fmt.Sprintf(violatorsQueryPattern, ignoredUsersPlaceholders, brokerDBName)
	return newRepo(query, ignoredUsers, db, logger, "quota violator")
}
