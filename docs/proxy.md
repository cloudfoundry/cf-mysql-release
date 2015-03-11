# Proxy

## All new connections are routed to the same node ##

The current proxy implementation is [Switchboard](https://github.com/cloudfoundry-incubator/switchboard). It proxies TCP connections between the client and nodes of the MariaDB Galera Cluster. All connections will be routed to a single active node; when that node fails the proxy should fail over to a different node. The proxy is configured to behave in this manner out of the box.

It should be noted that in the current configuration, it is only guaranteed that a **single** proxy node will behave this way. When we deploy multiple proxy nodes, there is a small probability that multiple proxy instances will route connections to different nodes and violate the desired invariant that all connections are routed to the same MariaDB node. This is a known issue, and we are currently exploring methods to circumvent it.

## Connection handling with Healthcheck

The proxy queries an HTTP healthcheck process, co-located on the database node, when determining where to route traffic. If the healthcheck process returns HTTP status code of 200, the node is considered healthy. In the case of failover, it will be considered as a candidate for new connections. If the healthcheck returns HTTP status code 503, the node is considered unhealthy. Clients with existing connections to a newly-unhealthy database node will find the connection severed, and are expected to make a reconnect attempt. At this point the proxy will route this new connection to a healthy node, assuming such a node exists.

## Connection handling on MariaDB failure ##

### MariaDB process on a node dies ###

The node is removed from the pool of healthy nodes. Any existing connections are severed; all new connections (and reconnections) are routed to another healthy node.

### A previously dead MariaDB node is resurrected ###

The resurrected node will not immediately receive connections, but is added to the pool of healthy nodes after it has reached a **synced** state. The proxy will continue to route all connections, new or existing, to the currently active node.

## Connection handling during State Snapshot Transfer (SST)

When a new node is added to the cluster it gets its state from an existing node via a process called SST. Because SST is performed by xtrabackup, the donor node continues to allow both reads and writes during this process.

## Connection handling for non-primary components ##

If a cluster loses more than than half its nodes, the remaining nodes lose quorum and form a **non-primary component**. It is also possible an individual node to become non-primary if it is unable to connect to greater than half of the cluster due to a network partition. In all cases, there is a six second grace period during which the cluster acknowledges something is wrong and gives missing nodes a chance to rejoin.

During the grace period, existing connections are maintained and new connections can be established. Read requests are fulfilled but write requests are suspended (requests hang).

Once the 6 second grace period expires, nodes found in a **primary component** will return to normal function, fulfilling write requests. Connections to nodes found in a **non-primary component** will be severed; new connections will routed to a healthy node.

# Proxy API


### Proxy API

The proxy hosts a JSON API at `proxy-<bosh job index>.p-mysql.<system domain>:<api_port>/v0/`. By default, this API is available on port 80.

Request:
*  Method: GET
*  Path: `/v0/backends`
*  Params: ~
*  Headers: Basic Auth

Response:

```
[
  {
    "name": "mysql-0",
    "ip": "1.2.3.4",
    "healthy": true,
    "active": true,
    "currentSessionCount": 2
  },
  {
    "name": "mysql-1",
    "ip": "5.6.7.8",
    "healthy": false,
    "active": false,
    "currentSessionCount": 0
  },
  {
    "name": "mysql-2",
    "ip": "9.9.9.9",
    "healthy": true,
    "active": false,
    "currentSessionCount": 0
  }
]
```

# Setting up Round Robin DNS with the proxy

## AWS Route 53

To set up a Round Robin DNS across multiple proxy IPs using AWS Route 53,
follow the following instructions:

1. Log in to AWS.
2. Click Route 53.
3. Click Hosted Zones.
4. Select the hosted zone that contains the domain name to apply round robin routing to.
5. Click 'Go to Record Sets'.
6. Select the record set containing the desired domain name.
7. In the value input, enter the IP addresses of each proxy VM, separated by a newline.

# Setting up a Load Balancer with multiple proxies

The proxy tier is responsible for routing connections from applications to healthy MariaDB cluster nodes, even in the event of node failure.

Bound applications are provided with a hostname or IP address to reach a database managed by the service. By default, the MySQL service will provide bound applications with the IP of the first instance in the proxy tier. Even if additional proxy instances are deployed, client connections will not be routed through them. This means the first proxy instance is a single point of failure.

**In order to eliminate the first proxy instance as a single point of failure, operators must configure a load balancer to route client connections to all proxy IPs, and configure the MySQL service to give bound applications a hostname or IP address that resolves to the load balancer.**

#### Configuring load balancer

To load balance across the proxies, configure the load balancer with each of the proxy IPs, and direct MySQL TCP traffic to port 3306. To use a health checker or monitor to test the health of each proxy, configure your load balancer to try TCP against the proxy health port (default is 1936).


#### Configuring cf-mysql-release to use load balancer
To ensure that bound applications will use the load balancer, the cf-mysql-broker job in the BOSH manifest must be changed:

```
jobs:
- name: cf-mysql-broker
  properties:
    mysql_node:
      host: <load balancer address>
```
