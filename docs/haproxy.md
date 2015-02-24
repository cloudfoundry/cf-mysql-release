# HAProxy and failover

## All new connections are routed to the same node ##

HAProxy is an application which proxies tcp connections and http requests. We use it to route connections to nodes of the MariaDB Galera Cluster. Preferably, all connections should be routed to a single node; when that node fails, the proxy should fail over to a different node. HAProxy supports this fail-over behavior with the following configuration:

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

It should be noted that this configuration only guarantees that a **single** HAProxy node will behave this way. When we deploy multiple HAProxy nodes, there is a small probably that the two proxies will fail-over to different nodes and violate the desired invariant that all connections are routed to the same Mariadb node. This is a known issue, and we are currently exploring methods to circumvent it.

## Connection handling with Healthcheck

HAProxy does not exclusively use the healthcheck when determining where to route traffic. If HAProxy attempts to establish a connection to a node where the healthcheck returns true, and that connection attempt fails, HAProxy will ignore the healthcheck and failover to a new node.

## Connection handling on MariaDB failure ##

The observations below are verifications of use cases only where connections are dropped due to the MariaDB process dying.

### MariaDB process on a node dies ###

The node is removed from the pool of healthy nodes. Any existing connections are dropped; all new connections (and reconnections) are routed to another healthy node.

### A previously dead MariaDB node is resurrected ###

The resurrected node will not receive connections. HAProxy has already failed-over to a new node; all connections, new or existing, will go to that node instead.

### Untested ###

What happens if a node dies and is resurrected between ping intervals? Perhaps HAProxy routes traffic to bad nodes and application see multiple connection failures before node becomes alive again. Mitigated if we reduce ping interval (see further discussion below).

## Connection handling during State Snapshot Transfer (SST)

When a new node is added to the cluster it gets its state from an existing node via a process called SST. Because SST is performed by xtrabackup, the donor node continues to allow both reads and writes during this process.

## Connection handling for non-primary components ##

If a cluster loses more than than half its nodes then the remaining nodes form a non-primary component. There is a currently a six second grace period during which the cluster acknowledges something is wrong and gives missing nodes a chance to rejoin.

During the grace period, existing connections are maintained and new connections can be established. Read requests are fulfilled but write requests are suspended (requests hang).

Once the 6 second grace period expires, nodes in primary component will return to normal function, fulfilling write requests.

Upon expiry of the grace period, nodes in a non-primary component will maintain existing connections and new connections can be established, but nearly all requests will receive the error `WSREP has not yet prepared this node for application use`, prompting the client to close the connection. Clients with open connections and hung writes will immediately receive this error and are expected to close the connection once the nodes enter a non-primary component.

# Setting up Round Robin DNS with HAProxy

## AWS Route 53

To set up a Round Robin DNS across multiple HAProxy IPs using AWS Route 53,
follow the following instructions:

1. Log in to AWS.
2. Click Route 53.
3. Click Hosted Zones.
4. Select the hosted zone that contains the domain name to apply round robin routing to.
5. Click 'Go to Record Sets'.
6. Select the record set containing the desired domain name.
7. In the value input, enter the IP addresses of each HAProxy VM, separated by a newline.

# Setting up a Load Balancer with multiple proxies

The proxy tier is responsible for routing connections from applications to healthy MariaDB cluster nodes, even in the event of node failure.

Bound applications are provided with a hostname or IP address to reach a database managed by the service. By default, the MySQL service will provide bound applications with the IP of the first instance in the proxy tier. Even if additional proxy instances are deployed, client connections will not be routed through them. This means the first proxy instance is a single point of failure.

**In order to eliminate the first proxy instance as a single point of failure, operators must configure a load balancer to route client connections to all proxy IPs, and configure the MySQL service to give bound applications a hostname or IP address that resolves to the load balancer.**

#### Configuring load balancer

To load balance across the proxies, configure the load balancer with each of the proxy IPs, and direct traffic to port 3306. To use a health checker or monitor to test the health of each proxy, configure your load balancer to try TCP against port 1936.


#### Configuring cf-mysql-release to use load balancer
To ensure that bound applications will use the load balancer, the cf-mysql-broker job in the BOSH manifest must be changed:

```
jobs:
- name: cf-mysql-broker
  properties:
    mysql_node:
      host: <load balancer address>
```

# Known Issues #

In states such as SST and Non-primary Components (see above), MariaDB is operational, disallows writes, but does not terminate connections.

HAProxy only considers the galera-healthcheck (available on port 9200 of each node) in determining where to route **new** (as opposed to existing) connections. In the aforementioned states, the healthcheck will report a node as unhealthy, so new connections will be routed to a healthy node, but it is not a feature of HAProxy to terminate existing connections.

Depending on the timeout configured by the client, HAProxy's failure to server existing connections could cause long wait times for clients connected to unhealthy nodes. The team is currently working on solutions to the problem.

## Further Discussion ##

* Pinging interval - should it be faster? HAProxy is efficient so increasing the pinging frequency might result in better responsiveness without sacrificing performance.
