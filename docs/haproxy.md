---
title: HAProxy
---

HAProxy is an application which provides load-balancing and high-availability. We use it to route connections between multiple MariaDB nodes. The high-availability feature of HAProxy is used to failover between nodes, but the load-balance feature is not used, as we want to ensure that all connections (read and write) go to a single node. To achieve this, we use the HAProxy config as follows:

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

## Experiments and Results ##

For the purpose of this section, all failures are MariaDB failures, not galera-healthcheck failures (see known issues below)

### Spawning new connections ###

No matter the number of healthy nodes, all connections go to the same node.

### MariaDB on a node dies with no existing connections ###

The node is removed from the pool of healthy nodes. New connections route to another healthy node.

### MariaDB on a node dies with existing connections ###

Existing connections are cut, and on the next connection attempt they are routed to the same healthy node as all other connections.

### MariaDB node is resurrected with no connections on other nodes ###

HAProxy has already failed-over to a new node, so all connections will go to that.

### MariaDB node is resurrected with existing connections on other nodes ###

As above, all connections are already routed to another node so this doesn't affect the routing of new or existing connections.

## Untested ##

* What happens if a node dies and is resurrected between ping intervals? Perhaps HAProxy routes traffic to bad nodes and application see multiple connection failures before node becomes alive again. Mitigated if we reduce ping interval (see further discussion below).

## Known Issues ##

* HAProxy only checks port 9200 (galera-healthcheck) when considering new connections - existing connections will continue to be routed to the MariaDB node if it is alive. This may cause an issue if the node goes into DONOR mode as HAProxy will not automatically reroute existing connections. This issue will likely be resolved when we rewrite the galera-healthcheck to be a proxy for the connection, as it will be responsible for severing the connection for all states and HAProxy will not have to perform a healthcheck on a separate port.

## Further Discussion ##

* Pinging interval - should it be faster? HAProxy is efficient so increasing the pinging frequency might result in better responsiveness without sacrificing performance.
