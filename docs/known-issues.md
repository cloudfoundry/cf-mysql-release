### Build issue with cf-mysql v26

The release tagged `v26` will no longer build from source due to a dependency that is no longer available. To continue working from source, update your source tree to `develop`.

###MyISAM Tables
The clustering plugin used in this release (Galera) does not support replication of MyISAM Tables. However, the service does not prevent the creation of MyISAM tables. When MyISAM tables are created, the tables will be created on every node (DDL statements are replicated), but data written to a node won't be replicated. If the persistent disk is lost on the node data is written to (for MyISAM tables only), data will be lost. To change a table from MyISAM to InnoDB, please follow this [guide](http://dev.mysql.com/doc/refman/5.5/en/converting-tables-to-innodb.html).

###Max User Connections
When updating the max_user_connections property for an existing plan, the connections currently open will not be affected (ie if you have decreased from 20 to 40, users with 40 open connections will keep them open). To force the changes upon users with open connections, an operator can restart the proxy job as this will cause the connections to reconnect and stay within the limit.  Otherwise, if any connection above the limit is reset it won't be able to reconnect, so the number of connections will eventually converge on the new limit.
