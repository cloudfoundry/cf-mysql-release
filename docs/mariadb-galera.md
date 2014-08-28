---
title: MariaDB Galera Cluster
---

## Connection behavior during state transfer

When a new node is added to the cluster it gets its state from an existing node via a process called SST.  During this process, the donor node suspends writes but allows reads.  It holds open existing connections and also allows new connections.  It doesn't return an error on write, but instead writes hang until the SST is over.

### Untested ###

* As writes to DONOR node are suspended during SST, it is conceivable that the connection may time out if the SST takes a long time. We have not managed to reproduce this, but it might be possible to observe this behavior if the cluster is running for a long time before adding a new node. This will not be an issue when we implement a proxy on the node, as this will sever the connection as soon as MariaDB enters DONOR mode.

## Connection behavior for non-primary component ##

If a cluster loses n/2 + 1 nodes (i.e. just over half of the nodes) then the remaining nodes form a non-primary component. In this state it is impossible to perform any meaningful operations - reads and writes are met with an error - `WSREP has not yet prepared this node for application use`. It is possible to perform some admin tasks e.g. `show databases`. Even `use database xyz` failed. Existing connections behave the same as new connections - everything is met with the same error.
