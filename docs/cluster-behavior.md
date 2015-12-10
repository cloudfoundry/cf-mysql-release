# Cluster scaling, node failure, and quorum

Documented here are scenarios in which the size of a cluster may change, how the cluster behaves, and how to restore service function when impacted. [Galera Cluster](http://galeracluster.com) is used to manage the [MariaDB](https://mariadb.com/kb/en/mariadb/what-is-mariadb-galera-cluster/) cluster in our release.

### Healthy Cluster

Galera documentation refers to nodes in a healthy cluster as being part of a [primary component](http://galeracluster.com/documentation-webpages/glossary.html#term-primary-component). These nodes will respond normally to all queries, reads, writes and database modifications.

If an individual node is unable to connect to the rest of the cluster (ex: network partition) it becomes non-primary (stops accepting writes and database modifications). In this case, the rest of the cluster should continue to function normally. A non-primary node may eventually regain connectivity and rejoin the primary component.

If more than half of the nodes in a cluster are no longer able to connect to each other, all of the remaining nodes lose quorum and become non-primary. In this case, the cluster must be manually restarted, as documented in the [bootstrapping docs](bootstrapping.md).

### Graceful removal of a node
  - Shutting down a node with monit (or decreasing cluster size by one) will cause the node to gracefully leave the cluster.
  - Cluster size is reduced by one and maintains healthy state. Cluster will continue to operate, even with a single node, as long as other nodes left gracefully.

### Adding new nodes

When new nodes are added to or removed from a MySQL service, a top-level property is updated with the new nodes' IP addresses. As BOSH deploys, it will update the configuration and restart all of the mysql nodes **and** the proxy nodes (to inform them of the new IP addresses as well). Restarting the nodes will cause all connections to that node to be dropped while the node restarts.

### Scaling the cluster

#### Scaling up from 1 to N nodes
When a new MariaDb node comes online, it replicates data from the existing node in the cluster. Once replication is complete, the node will join the cluster. The proxy will continue to route all incoming connections to the primary node while it remains healthy.

If the proxy detects that this node becomes [unhealthy](https://github.com/cloudfoundry/cf-mysql-release/blob/release-candidate/docs/proxy.md#unhealthy), it will sever existing connections, and route all new connections to a different, healthy node. If there are no healthy MariaDb nodes, the proxy will reject all subsequent connections.

While transitioning from one node to a cluster, there will be an undetermined period of performance degradation while the new node sync's all data from the original node.

Note: If you are planning to scale up MariaDb nodes, it is recommended to do so in different Availability Zones to maximize cluster availability. An Availability Zone is a network-distinct section of a given Region. Further details are available in [Amazon's documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html).

#### Scaling down from N to 1 node
When scaling from multiple nodes to a single MariaDb node, the proxy will determine that the sole remaining node is the primary node (provided it remains healthy). The proxy routes incoming connections to the remaining MariaDb node.

### Rejoining the cluster (existing nodes)
Existing nodes restarted with monit should automatically join the cluster. If an existing node fails to join the cluster, it may be because its transaction record's (`seqno`) is higher than that of the nodes in the cluster with quorum (aka the primary component).

  - If the node has a higher `seqno` it will be apparent in the error log `/var/vcap/sys/log/mysql/mysql.err.log`.
  - If the healthy nodes of a cluster have a lower transaction record number than the failing node, it might be desirable to shut down the healthy nodes and bootstrap from the node with the more recent transaction record number. See the [bootstraping docs](bootstrapping.md) for more details.
  - Manual recovery may be possible, but is error-prone and involves dumping transactions and applying them to the running cluster (out of scope for this doc).
  - Abandoning the data is also an option, if you're ok with losing the unsynced transactions. Follow the following steps to abandon the data (as root):
    - Stop the process with `monit stop mariadb_ctrl`.
    - Delete the galera state (`/var/vcap/store/mysql/grastate.dat`) and cache (`/var/vcap/store/mysql/galera.cache`) files from the persistent disk.
    - Restarting the node with `monit start mariadb_ctrl`.

### Quorum
  - In order for the cluster to continue accepting requests, a quorum must be reached by peer-to-peer communication. More than half of the nodes must be responsive to each other to maintain a quorum.
  - If more than half of the nodes are unresponsive for a period of time the nodes will stop responding to queries, the cluster will fail, and bootstrapping will be required to re-enable functionality.

### Avoid an even number of nodes
  - It is generally recommended to avoid an even number of nodes. This is because a partition could cause the entire cluster to lose quorum, as neither remaining component has more than half of the total nodes.
  - A 2 node cluster cannot tolerate the failure of single node failure as this would cause loss of quorum. As such, the minimum number of nodes required to tolerate single node failure is 3.

### Unresponsive node(s)
  - A node can become unresponsive for a number of reasons:
    - network latency
    - mysql process failure
    - firewall rule changes
    - vm failure
  - Unresponsive nodes will stop responding to queries and, after timeout, leave the cluster.
  - Nodes will be marked as unresponsive (innactive) either:
    - If they fail to respond to one node within 15 seconds
    - OR If they fail to respond to all other nodes within 5 seconds
  - Unresponsive nodes that become responsive again will rejoin the cluster, as long as they are on the same IP which is pre-configured in the gcomm address on all the other running nodes, and a quorum was held by the remaining nodes.
  - All nodes suspend writes once they notice something is wrong with the cluster (write requests hang). After a timeout period of 5 seconds, requests to non-quorum nodes will fail. Most clients return the error: `WSREP has not yet prepared this node for application use`. Some clients may instead return `unknown error`. Nodes who have reached quorum will continue fulfilling write requests.
  - If deployed using a proxy, a continually inactive node will cause the proxy to fail over, selecting a different mysql node to route new queries to.

### Re-bootstrapping the cluster after quorum is lost
  - The start script will currently bootstrap node 0 only on initial deploy. If bootstrapping is necessary at a later date, it must be done manually. For more information on bootstrapping a cluster, see [Bootstrapping Galera](bootstrapping.md).
  - If the single node is bootstrapped, it will create a new one-node cluster that other nodes can join.

### Simulating node failure
  - To simulate a temporary single node failure, use `kill -9` on the pid of the mysql process. This will only temporarily disable the node because the process is being monitored by monit, which will restart the process if it is not running.
  - To more permenantly disable the process, execute `monit unmonitor mariadb_ctrl` before `kill -9`.
  - To simulate multi-node failure without killing a node process, communication can be severed by changing the iptables config to dissallow communication:

    ```
    iptables -F && # optional - flush existing rules \
    iptables -A INPUT -p tcp --destination-port 4567 -j DROP && \
    iptables -A INPUT -p tcp --destination-port 4568 -j DROP && \
    iptables -A INPUT -p tcp --destination-port 4444 -j DROP && \
    iptables -A INPUT -p tcp --destination-port 3306 && \
    iptables -A OUTPUT -p tcp --destination-port 4567 -j DROP && \
    iptables -A OUTPUT -p tcp --destination-port 4568 -j DROP && \
    iptables -A OUTPUT -p tcp --destination-port 4444 -j DROP && \
    iptables -A OUTPUT -p tcp --destination-port 3306
    ```

    To recover from this, drop the partition by flushing all rules:
    ```
    iptables -F
    ```
