# Cluster Configuration

This page documents the various configuration decisions that have been made in relation to MariaDB and Galera in cf-mysql-release.

### SST method

Galera supports multiple methods for [State Snapshot Transfer](http://www.percona.com/doc/percona-xtradb-cluster/5.5/manual/state_snapshot_transfer.html).
The `rsync` method is usually fastest. The `xtrabackup` method has the advantage of keeping the donor node writeable during SST. We have chosen to use `xtrabackup`.

### InnoDB Redo Log Files

Our cluster defaults to 1GB for log file size to support large blob transactions. By default there are two log files, and that number is static throughout the life of the database. Thus, the minimum disk size required for a DB install is 2GB per node, separate from the data storage required. To learn more about redologs, check out [redo logging in innodb](https://blogs.oracle.com/mysqlinnodb/entry/redo_logging_in_innodb).

### Max User Connections
To ensure all users get fair access to system resources, we have capped each user's number of connections to 40.

### Skip External Locking
Since each Virtual Machine only has one mysqld process running, we do not need external locking.

### Max Allowed Packet
We allow blobs up to 256MB. This size is unlikely to limit a user's query, but is also manageable for our InnoDB log file size. Consider using our Riak CS service for storing large files.

### Innodb File Per Table
Innodb allows using either a single file to represent all data, or a separate file for each table. We chose to use a separate file for each table as this provides more flexibility and optimization. For a full list of pros and cons, see MySQL's documentation for [InnoDB File-Per-Table Mode](http://dev.mysql.com/doc/refman/5.5/en/innodb-multiple-tablespaces.html).

### Innodb File Format
To take advantage of all the extra features available with the `innodb_file_per_table = ON` option, we use the `Barracuda` file format.

### Temporary Tables

MySQL is configured to convert temporary in-memory tables to temporary on-disk tables when a query EITHER generates more than 16 million rows of output or uses more than 32MB of data space.
Users can see if a query is using a temporary table by using the EXPLAIN command and looking for "Using temporary," in the output.
If the server processes very large queries that use /tmp space simultaneously, it is possible for queries to receive no space left errors.
