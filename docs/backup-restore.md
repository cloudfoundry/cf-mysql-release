# Backup and Restore

## Backing Up

### All Databases

If remote admin access is enabled, use [mysqldump](https://mariadb.com/kb/en/mariadb/mysqldump/) to back up the data as the admin user.

<p class="note"><strong>Note</strong>: This backup acquires a global read lock on all tables, but does not hold it for the entire duration of the dump.</p>

To back up all databases in the MySQL deployment, run the following command:

> mysqldump -u admin -p PASSWORD -h MYSQL-NODE-IP --all-databases --single-transaction > BACKUP-NAME.sql

Where:

* `PASSWORD`: Enter the admin password from the manifest.
* `MYSQL-NODE-IP`: Enter the MySQL node IP address from `bosh -d cf-mysql instances`.
* `BACKUP-NAME`: Enter a name for the backup file.

For example:
```
$ mysqldump -u admin -p 123456789 \
  -h 10.10.10.8 --all-databases \
  --single-transaction > user_databases.sql
```

### Single Service Instance

To back up a single service instance, run the following command:

> mysqldump -u admin -p PASSWORD -h MYSQL-NODE-IP DB-NAME --single-transaction > BACKUP-NAME.sql


Where, in addition to above:
* `DB-NAME`: Enter the name of the database you want to back up.

For example:
```
$ mysqldump -u admin -p 123456789 \
-h 10.10.10.8 cf_2033da4e_d0c8_4a12_9130_b349473f9fac \
--single-transaction > user_databases.sql
```

## Restoring

The procedure for restoring your MySQL data from a manual backup requires more steps when restoring a backup of all user databases. Executing the SQL dump will drop, recreate, and refill the specified databases and tables.

**WARNING:** Restoring a database deletes all data that existed in the database before the restore. Restoring a database using a full backup artifact, produced by <code>mysqldump --all-databases</code> for example, replaces all data and user permissions.

For a point in time recovery see the [mysql docs](https://dev.mysql.com/doc/refman/5.7/en/point-in-time-recovery.html).

Perform the following steps to restore your MySQL data from a manual backup with remote admin access enabled:

1. When restoring all databases, if running in HA configuration, reduce the size of the cluster to a single node, and redeploy.
1. When restoring all databases, use the MySQL client to enable the creation of tables using any storage engine. Run the following command:
    > mysql -u admin -p PASSWORD -h MYSQL-NODE-IP -e "SET GLOBAL enforce_storage_engine=NULL"
    > mysql -u admin -p PASSWORD -h MYSQL-NODE-IP -e "SET GLOBL slow_query_log=OFF"

    Where:
    * `PASSWORD`: Enter the admin password from the manifest.
    * `MYSQL-NODE-IP`: Enter the MySQL node IP address from `bosh instances`.

    For example:
    ```
    $ mysql -u abcdefghijklm -p 123456789 \
    -h 10.10.10.8 -e "SET GLOBAL enforce_storage_engine=NULL"
    $ mysql -u abcdefghijklm -p 123456789 \
    -h 10.10.10.8 -e "SET GLOBL slow_query_log=OFF"
    ```

1. When restoring all databases, **or** just a single instance, use the MySQL client to restore the MySQL database or databases. Run the following command:

    > mysql -u admin -p PASSWORD -h MYSQL-NODE-IP < BACKUP-NAME.sql

    Where:
    * `BACKUP-NAME`: Enter the file name of the backup artifact.

    For example:
    ```
    $ mysql -u admin -p 123456789 \
    -h 10.10.10.8 -e < user_databases.sql
    ```

1. When restoring all databases, use the MySQL client to restore the original storage engine restriction. Run the following command:

    > mysql -u admin -p PASSWORD -h MYSQL-NODE-IP -e "SET GLOBAL enforce_storage_engine='InnoDB'"

    And optionally:
    > mysql -u admin -p PASSWORD -h MYSQL-NODE-IP -e "SET GLOBAL slow_query_log=ON"

    For example:
    ```
    $ mysql -u admin -p 123456789 \
    -h 10.10.10.8 -e "SET GLOBAL enforce_storage_engine='InnoDB'"
    ```

1. When restoring all databases, if you are running in HA mode, re-configure the cluster to use three nodes and redeploy. If you are not running HA mode, restart the database server. This step is not necessary if scaling back to three MySQL nodes.
    ```
    $ monit stop mariadb_ctrl
    $ monit start mariadb_ctrl
    ```
