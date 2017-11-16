# Cluster Configuration

This page documents the various configuration decisions that have been made
in relation to MariaDB and Galera in cf-mysql-release.

### Maximum Open Files

In the course of normal operation, where MySQL VMs enter and leave the cluster,
there are times when a node will need to open file descriptors to all of the
database files. The default is 65,000. You can increase this by including a
different value for `cf_mysql.mysql.max_open_files` in the manifest, and re-deploying.

### Slow Query Log

cf-mysql automatically enables the
[slow query log](https://mariadb.com/kb/en/mariadb/slow-query-log-overview/)
and sets it to record any query that takes longer than 10 seconds.
By default, those logs appear in the file
`/var/vcap/sys/log/mysql/mysql_slow_query.log` per node.
For a consolidated view, use a syslog server.

### SST method

Galera supports multiple methods for
[State Snapshot Transfer](http://www.percona.com/doc/percona-xtradb-cluster/5.5/manual/state_snapshot_transfer.html).
We have chosen the `xtrabackup` method, as it has the advantage of keeping the
donor node writeable during SST.

### InnoDB Redo Log Files

Our cluster defaults to 1GB for log file size to support large blob transactions.
By default there are two log files, and that number is static throughout the
life of the database. Thus, the minimum disk size required for a DB install is
2GB per node, separate from the data storage required. To learn more about
redologs, check out
[redo logging in innodb](https://blogs.oracle.com/mysqlinnodb/entry/redo_logging_in_innodb).

### Max User Connections

To ensure all users get fair access to system resources, we have capped each
user's number of connections to 40.

### Skip External Locking

Since each Virtual Machine only has one mysqld process running, we do not need
external locking.

### Max Allowed Packet

We allow blobs up to 256MB. This size is unlikely to limit a user's query,
but is also manageable for our InnoDB log file size.

### Query Cache Type and Size

Matching MariaDB's default as of [version 10.1.7](https://mariadb.com/kb/en/mariadb/mariadb-1017-release-notes/) and later, cf-mysql-release disables `query_cache_type` and set `query_cache_size = 0`.

### Innodb Buffer Pool Size

You can tune the InnoDB Buffer Pool size via two different properties. The absolute number of bytes to allocate to the buffer pool can be specified by the `cf_mysql.mysql.innodb_buffer_pool_size` property. If you deploy to a variety of VM sizes, you can specify a percentage of total system memory using the `cf_mysql.mysql.innodb_buffer_pool_size_percent` property instead. The buffer pool will be re-calculated on every deploy. If both properties are specified, the absolute number of bytes,`cf_mysql.mysql.innodb_buffer_pool_size`, is used.

If deploying to bosh lite, use `cf_mysql.mysql.innodb_buffer_pool_size`. The percentage calculation will
be based on the total bosh lite VM size and will cause the 3 nodes to consume more memory than desired.

### Innodb File Per Table

Innodb allows using either a single file to represent all data, or a separate
file for each table. We chose to use a separate file for each table as this
provides more flexibility and optimization. For a full list of pros and cons,
see MySQL's documentation for
[InnoDB File-Per-Table Mode](http://dev.mysql.com/doc/refman/5.5/en/innodb-multiple-tablespaces.html).

### Innodb File Format

To take advantage of all the extra features available with the
`innodb_file_per_table = ON` option, we use the `Barracuda` file format.

### Innodb Flush Method

Defines the method used to flush data to InnoDB data files, which can affect I/O throughput. If unspecified, fsync() is
used to flush data and logs. Buffered I/O is used when reading from disk and uses the OS caching mechanisms.
`cf_mysql.mysql.innodb_flush_method` may be set to `O_DIRECT` to avoid double-buffering, which may be more performant
when `innodb_buffer_pool_size` is also set sufficiently large. `O_DIRECT` does not use the OS file caching mechanisms
and may be slower in some environments.

### Innodb Strict Mode

Innodb Strict Mode is disabled by default by can be enabled by configuring the
`cf_mysql.mysql.innodb_strict_mode` property of the mysql job to `true`

### Temporary Tables

MySQL is configured to convert temporary in-memory tables to temporary on-disk
tables when a query EITHER generates more than 16 million rows of output or
uses more than 32MB of data space. Users can see if a query is using a temporary
table by using the EXPLAIN command and looking for "Using temporary," in the output.
If the server processes very large queries that use /tmp space simultaneously,
it is possible for queries to receive no space left errors.

### Security Concerns

MySQL as its associated processes are run as `vcap`, never `root`.

cf-mysql always uses the `--skip-symbolic-links` flag to prevent use of symbolic links.
This is a general security recommendation for MySQL
([docs](https://dev.mysql.com/doc/refman/5.7/en/security-against-attack.html)). With
symbolic links enabled, if somebody had write access to the data directory, they could
change or delete other files owned by the same user.

We expose an `enable_local_file` property with a default of `true` to toggle the
`local_infile` system variable. If this is disabled, the server will refuse all LOAD DATA LOCAL
statements, punting those with an ERROR 1148.

`secure_file_priv` has been configured to `/var/vcap/data/mysql/files` to ensure
files are not accidentally exposed.

The following default users are created:

| User | Permissions |
|---|---|
| cf_mysql.mysql.admin_username@localhost | ALL PRIVILEGES ON \*.\* WITH GRANT OPTION |
| cluster-health-logger@127.0.0.1 | USAGE ON \*.\* |
| galera-healthcheck@127.0.0.1 | USAGE ON \*.\* |
| cf-mysql-broker@% | SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, REFERENCES, INDEX, ALTER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE ON \*.\* WITH GRANT OPTION |
| quota-enforcer@% | ALL PRIVILEGES ON \*.\* WITH GRANT OPTION |

`cf_mysql.mysql.admin_username@localhost` is only able to connect locally from the mysql nodes when `cf_mysql.mysql.remote_admin_access` is disabled.
