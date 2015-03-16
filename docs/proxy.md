# Proxy

In cf-mysql-release, [Switchboard](https://github.com/cloudfoundry-incubator/switchboard) is used to proxy TCP connections to healthy MariaDB nodes.

## Consistent Routing

At any given time, Switchboard will only route to one active node. That node will continue to be the only active node until it becomes unhealthy. 

If multiple Switchboard proxies are used in parallel (ex: behind a load-balancer) there is no guarantee that the proxies will choose the same active node. There is a small probability that multiple proxy instances will route connections to different nodes, which can cause race conditions on writes, leading to desycnhronized states and lost data. This is a known issue, with exploration of mitigation options on the roadmap. To completely avoid this problem, use a single proxy instance or an external failover system to switch between multiple switchboard instances.

## Node Health

### Healthy

The proxy queries an HTTP healthcheck process, co-located on the database node, when determining where to route traffic. 

If the healthcheck process returns HTTP status code of 200, the node is added to the pool of healthy nodes. 

A resurrected node will not immediately receive connections. The proxy will continue to route all connections, new or existing, to the currently active node. In the case of failover, all healthy nodes will be considered as candidates for new connections. 

### Unhealthy

If the healthcheck returns HTTP status code 503, the node is considered unhealthy. 

This happens when a node becomes non-primary, as specified by the [cluster-behavior docs](cluster-behavior.md).

The proxy will sever all existing connections to newly unhealthy nodes. Clients are expected to handle reconnecting on connection failure. The proxy will route new connections to a healthy node, assuming such a node exists.

### Unresponsive

If node health cannot be determined due to an unreachable or unresponsive healthcheck endpoint, the proxy will consider the node unhealthy. This may happen if there is a network partition or if the VM containing the healthcheck and MariaDB node died.


## State Snapshot Transfer (SST)

When a new node is added to the cluster or rejoins the cluster, it synchronizes state with the primary component via a process called SST. A single node from the primary component is chosen to act as a state donor. By default Galera uses rsync to perform SST, which blocks for the duration of the transfer. However, cf-mysql-release is configured to use [Xtrabackup](http://www.percona.com/doc/percona-xtrabackup), which allows the donor node to continue to accept reads and writes.


## Removing the proxy as a SPOF

The proxy tier is responsible for routing connections from applications to healthy MariaDB cluster nodes, even in the event of node failure.

Bound applications are provided with a hostname or IP address to reach a database managed by the service. By default, the MySQL service will provide bound applications with the IP of the first instance in the proxy tier. Even if additional proxy instances are deployed, client connections will not be routed through them. This means the first proxy instance is a single point of failure.

**In order to eliminate the first proxy instance as a single point of failure, operators must configure a load balancer to route client connections to all proxy IPs, and configure the MySQL service to give bound applications a hostname or IP address that resolves to the load balancer.**

### Configuring load balancer

Configure the load balancer to route traffic for port 3306 to the IPs of all proxy instances on port 3306. Configure a healthcheck to use TCP on port 1936. The port proxy instances listen for healthcheck is configurable with manifest property `properties.proxy.health_port` for the proxy job.

### Configuring cf-mysql-release to give applications the address of the load balancer
To ensure that bound applications will use the load balancer to reach bound databases, the manifest property `properties.mysql_node.host` must be updated for the cf-mysql-broker job:

```
jobs:
- name: cf-mysql-broker
  properties:
    mysql_node:
      host: <load balancer address>
```

### AWS Route 53

To set up a Round Robin DNS across multiple proxy IPs using AWS Route 53,
follow the following instructions:

1. Log in to AWS.
2. Click Route 53.
3. Click Hosted Zones.
4. Select the hosted zone that contains the domain name to apply round robin routing to.
5. Click 'Go to Record Sets'.
6. Select the record set containing the desired domain name.
7. In the value input, enter the IP addresses of each proxy VM, separated by a newline.

Finally, update the manifest property `properties.mysql_node.host` for the cf-mysql-broker job, as described above.

## API

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

