# Cluster Configuration

This page documents the various configuration decisions that have been made in relation to MariaDB and Galera in cf-mysql-release.

###SST method

Galera supports multiple methods for [State Snapshot Transfer](http://www.percona.com/doc/percona-xtradb-cluster/5.5/manual/state_snapshot_transfer.html).
The `rsync` method is usually fastest. The `xtrabackup` method has the advantage of keeping the donor node writeable during SST. We have chosen to use `rsync`

###InnoDB Log Files
To Be Determined (EITHER kept logs small to reduce SST time OR used 1G log file size to support larger blobs)

###Max User Connections
To ensure all users get fair access to system resources, we have capped each user's number of connections to 40.

###Skip External Locking
Since each Virtual Machine only has one mysqld process running, we do not need external locking.

###Max Allowed Packet
We allow blobs up to 256MB. This size is unlikely to limit a user's query, but is also manageable for our InnoDB log file size. Consider using our Riak CS service for storing large files.

###Innodb File Per Table
Innodb allows using either a single file to represent all data, or a separate file for each table. We chose to use a separate file for each table as this provides more flexibility and optimization. For a full list of pros and cons, see MySQL's documentation for [InnoDB File-Per-Table Mode](http://dev.mysql.com/doc/refman/5.5/en/innodb-multiple-tablespaces.html).

###Innodb File Format
To take advantage of all the extra features available with the `innodb_file_per_table = ON` option, we use the `Barracuda` file format.
