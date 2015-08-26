# Show number of each plan type that has been allocated
SELECT max_storage_mb AS plan_size,
	   COUNT(db_name) AS total_instances_allocated
FROM mysql_broker.service_instances
GROUP BY max_storage_mb
ORDER BY max_storage_mb DESC;
