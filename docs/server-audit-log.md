## Server Audit Log

The purpose of the MariaDB Audit Plugin is to log the server's activity.

Records about who connected to the server, what queries ran and what tables were touched can be stored to the rotating log file or sent to the local syslogd.

## Enabling the Server Audit Log

By default, this plugin is disabled. To enable the server audit log, set the `server_audit_events` property in your cf-mysql manifest.

```yml
properties:
	mysql:
		server_audit_events: connect,query
```

Valid values are: `connect`, `query`, `table`, `query_ddl`, `query_dml`, and `query_dcl`.

See (MariaDB audit plugin documentation)[https://mariadb.com/kb/en/mariadb/about-the-mariadb-audit-plugin/] for more information about logging options.

## Viewing Output of the Server Audit Log

By default, the audit log will live in /var/vcap/store/mysql_audit_logs/mysql_server_audit.log.

## Syslog

Due to the fact that the audit log may contain sensitive application data, it is not sent to syslog.
