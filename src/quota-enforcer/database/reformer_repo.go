package database

import (
	"database/sql"
	"fmt"
	"strings"

	"code.cloudfoundry.org/lager"
)

// LEFT JOIN is required so that dropping all tables will restore write access
const reformersQueryPattern = `
SELECT reformers.name AS reformer_db, reformers.user AS reformer_user
FROM (
	SELECT violator_dbs.name, violator_dbs.user, tables.data_length, tables.index_length
	FROM   (
		SELECT DISTINCT table_schema as name, replace(substring_index(schema_privileges.grantee, '@', 1), "'", '') AS user
		FROM information_schema.schema_privileges
		LEFT JOIN %s.read_only_users
			ON read_only_users.grantee = schema_privileges.grantee COLLATE utf8_general_ci
		WHERE privilege_type IN ('SELECT', 'INSERT', 'UPDATE', 'CREATE')
		  AND replace(substring_index(schema_privileges.grantee, '@', 1), "'", '') NOT IN (%s)
		  AND read_only_users.id IS NULL
		GROUP BY schema_privileges.grantee, table_schema
		HAVING count(*) != 4
	) AS violator_dbs
	JOIN        %s.service_instances AS instances ON violator_dbs.name = instances.db_name COLLATE utf8_general_ci
	LEFT JOIN   information_schema.tables AS tables ON tables.table_schema = violator_dbs.name
	GROUP  BY   violator_dbs.user
	HAVING ROUND(SUM(COALESCE(tables.data_length + tables.index_length,0) / 1024 / 1024), 1) < MAX(instances.max_storage_mb)
) AS reformers
`

func NewReformerRepo(brokerDBName string, ignoredUsers []string, db *sql.DB, logger lager.Logger) Repo {
	ignoredUsersPlaceholders := strings.Join(strings.Split(strings.Repeat("?", len(ignoredUsers)), ""), ",")
	query := fmt.Sprintf(reformersQueryPattern, brokerDBName, ignoredUsersPlaceholders, brokerDBName)
	return newRepo(query, ignoredUsers, db, logger, "quota reformer")
}
