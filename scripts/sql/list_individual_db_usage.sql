# Show all databases with used vs allocated size
SELECT db AS instance,
	   mb_used AS used_space_in_mb,
	   mb_plan_size AS allocated_space_in_mb,
	   ROUND((mb_used / mb_plan_size) * 100, 2) as percent_used
FROM
(SELECT instances.db_name AS db,
	   COALESCE(ROUND(SUM(tables.data_length + tables.index_length) / 1024 / 1024, 2), 0) as mb_used,
	   MAX(instances.max_storage_mb) as mb_plan_size
FROM   mysql_broker.service_instances AS instances
LEFT JOIN information_schema.tables AS tables   ON tables.table_schema = instances.db_name COLLATE utf8_general_ci
GROUP  BY instances.db_name
ORDER BY mb_used DESC) data_usage;
