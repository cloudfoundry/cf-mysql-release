# HAProxy and failover

## All new connections are routed to the same node ##

HAProxy is an application which provides load-balancing and high-availability. We use it to route connections to nodes of the MariaDB Galera Cluster. The high-availability feature of HAProxy is used to failover between nodes, but the load-balance feature is not used, as we want to ensure that all connections (read and write) go to a single node. To achieve this, we use an HAProxy config as follows:

```
global
    log 127.0.0.1   local1 info
    daemon
    user vcap
    group vcap
    maxconn 64000

defaults
    log global
    timeout connect 30000ms
    timeout client 300000ms
    timeout server 300000ms

listen mysql-cluster
    stick-table type ip size 1
    stick on dst
    bind 0.0.0.0:3306
    option httpchk GET / HTTP/1.1\r\nHost:\ www
    mode tcp
    option tcplog
    server mysql-0 10.244.1.2:3306 check port 9200 inter 5000 rise 2 fall 1
    server mysql-0 10.244.1.6:3306 check port 9200 inter 5000 rise 2 fall 1
    server mysql-0 10.244.1.10:3306 check port 9200 inter 5000 rise 2 fall 1

listen stats :1936
    mode http
    stats enable
    stats uri /
    stats auth admin:password
```

## Connection handling on MariaDB failure ##

The observations below are verifications of uses cases only where connections are dropped due to the MariaDB process dying.

### MariaDB on a node dies with no existing connections ###

The node is removed from the pool of healthy nodes. All new connections are routed to another healthy node.

### MariaDB node is resurrected with no connections on other nodes ###

HAProxy has already failed-over to a new node. All connections will go to that node; the resurrected node will not receive connections.

### MariaDB on a node dies with existing connections ###

Existing connections are dropped; all new connections (and reconnections) are routed to another healthy node.

### MariaDB node is resurrected with existing connections on other nodes ###

As above, HAProxy has already failed-over to a new node. All connections will go to that node; the resurrected node will not receive connections.

### Untested ###

What happens if a node dies and is resurrected between ping intervals? Perhaps HAProxy routes traffic to bad nodes and application see multiple connection failures before node becomes alive again. Mitigated if we reduce ping interval (see further discussion below).

## Connection handling during State Snapshot Transfer (SST)

When a new node is added to the cluster it gets its state from an existing node via a process called SST.  During this process, the donor node suspends writes but allows reads. MariaDB holds open existing connections and also allows new connections. It doesn't return an error on write; instead writes hang until the SST is completed.

### Untested ###

As writes to DONOR node are suspended during SST, it is conceivable that the connection may time out if the SST takes a long time. We have not managed to reproduce this, but it might be possible to observe this behavior if the cluster is running for a long time before adding a new node. This will not be an issue when we implement a proxy on the node, as this will sever the connection as soon as MariaDB enters DONOR mode.

## Connection handling for non-primary components ##

If a cluster loses n/2 + 1 nodes (i.e. greater than half) then the remaining nodes form a non-primary component. In this state it is impossible to perform any meaningful operations - reads and writes are met with an error - `WSREP has not yet prepared this node for application use`. It is possible to perform some admin tasks e.g. `show databases`. Even `use database xyz` failed. Existing connections behave the same as new connections - everything is met with the same error.

## Known Issues ##

In states such as SST and Non-primary Components (see above), MariaDB is operational, disallows writes, but does not terminate connections.

HAProxy only considers the galera-healthcheck (available on port 9200 of each node) in determining where to route new connections. In the aforementioned states, the healthcheck will report a node as unhealthy, so new connections will be routed to a healthy node, but it is not a feature of HAProxy to terminate existing connections.

This results in connections on multiple nodes, which is undesireable due to the possibility for deadlocks. As we assume most appliations have not been designed to tolerate deadlocking, we will attempt to prevent this from happening.

The current plan is to implement a mechanism on each node responsible for severing existing connections when the healthcheck reports a node as unhealthy.

## Further Discussion ##

* Pinging interval - should it be faster? HAProxy is efficient so increasing the pinging frequency might result in better responsiveness without sacrificing performance.
