# list count of empty vs non-empty instances
SELECT SUM(CASE WHEN mb_used > 0 THEN 1 ELSE 0 END) AS non_empty_instances,
		SUM(CASE WHEN mb_used = 0 THEN 1 ELSE 0 END) AS empty_instances,
		COUNT(mb_used) AS total_instances
FROM (
	SELECT COALESCE(ROUND(SUM(tables.data_length + tables.index_length) / 1024 / 1024, 2), 0) as mb_used
	FROM   mysql_broker.service_instances AS instances
	LEFT JOIN information_schema.tables AS tables   ON tables.table_schema = instances.db_name COLLATE utf8_general_ci
	GROUP  BY instances.db_name
) data_usage;
