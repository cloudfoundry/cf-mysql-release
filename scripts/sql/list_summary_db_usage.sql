# Show summary of database usage
SELECT COUNT(db) AS total_instances_count,
	   ROUND(SUM(mb_used), 2) AS total_mb_used,
	   ROUND(SUM(mb_plan_size), 2) AS total_mb_allocated,
	   ROUND((SUM(mb_used) / SUM(mb_plan_size)) * 100, 2) as total_percent_used
FROM
(SELECT instances.db_name AS db,
	   COALESCE(ROUND(SUM(tables.data_length + tables.index_length) / 1024 / 1024, 2), 0) as mb_used,
	   MAX(instances.max_storage_mb) as mb_plan_size
FROM   mysql_broker.service_instances AS instances
LEFT JOIN information_schema.tables AS tables   ON tables.table_schema = instances.db_name COLLATE utf8_general_ci
GROUP  BY instances.db_name
ORDER BY mb_used DESC) data_usage;
