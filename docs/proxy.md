# Proxy

## All new connections are routed to the same node ##

The current proxy implementation is [Switchboard](https://github.com/pivotal-cf-experimental/switchboard). It proxies TCP connections between the client and nodes of the MariaDB Galera Cluster. Preferably, all connections should be routed to a single node; when that node fails the proxy should fail over to a different node. The proxy is configured to behave in this manner out of the box.

It should be noted that in the current configuration, it is only guaranteed that a **single** proxy node will behave this way. When we deploy multiple proxy nodes, there is a small probably that the two proxies will fail-over to different nodes and violate the desired invariant that all connections are routed to the same MariaDB node. This is a known issue, and we are currently exploring methods to circumvent it.

## Connection handling with Healthcheck

The proxy queries an http healthcheck process co-located on the database node when determining where to route traffic. If the healthcheck process returns http status code 200 the node is considered healthy and the proxy will route traffic to it. If the healthcheck returns http status code 503 the node is considered unhealthy and the proxy will not route new connections to it. Clients with existing connections to the newly-unhealthy database node will find the connection severed, and are expected to make a reconnect attempt. At this point the proxy will route this new connection to a healthy node, assuming such a node exists.

## Connection handling on MariaDB failure ##

The observations below are verifications of use cases only where connections are dropped due to the MariaDB process dying.

### MariaDB process on a node dies ###

The node is removed from the pool of healthy nodes. Any existing connections are severed; all new connections (and reconnections) are routed to another healthy node.

### A previously dead MariaDB node is resurrected ###

The resurrected node will not receive connections. The proxy has already failed-over to a new node; all connections, new or existing, will go to that node instead.

### Untested ###

What happens if a node dies and is resurrected between ping intervals? Perhaps the proxy routes traffic to bad nodes and applications see multiple connection failures before node becomes alive again. The ping interval is reasonably short so it's unlikely a node could fail and come back online before the proxy has severed connections and chosen a new healthy node (see further discussion below).

## Connection handling during State Snapshot Transfer (SST)

When a new node is added to the cluster it gets its state from an existing node via a process called SST. Because SST is performed by xtrabackup, the donor node continues to allow both reads and writes during this process.

## Connection handling for non-primary components ##

If a cluster loses more than than half its nodes then the remaining nodes form a non-primary component. There is a currently a six second grace period during which the cluster acknowledges something is wrong and gives missing nodes a chance to rejoin.

During the grace period, existing connections are maintained and new connections can be established. Read requests are fulfilled but write requests are suspended (requests hang).

Once the 6 second grace period expires, nodes in primary component will return to normal function, fulfilling write requests.

Upon expiry of the grace period, nodes in a non-primary component will maintain existing connections and new connections can be established, but nearly all requests will receive the error `WSREP has not yet prepared this node for application use`, prompting the client to close the connection. Clients with open connections and hung writes will immediately receive this error and are expected to close the connection once the nodes enter a non-primary component.

# Proxy API


### Proxy API

The proxy hosts a json api at `proxy-<bosh job index>.p-mysql.<system domain>:80/v0/`

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

# Known Issues #

In states such as SST and Non-primary Components (see above), MariaDB is operational, disallows writes, but does not terminate connections.

## Further Discussion ##

* Pinging interval - should it be faster? Currently the frequency at which the proxy polls the healthcheck is configurable via the manifest property `proxy.healthcheck_timeout_millis` as the polling frequency is a fixed fraction of the timeout.
