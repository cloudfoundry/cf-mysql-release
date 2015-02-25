# Bootstrapping a Galera Cluster

Bootstrapping is the process of (re)starting a Galera cluster.  

## When to Bootstrap

- Manual bootstrapping should only be required if all nodes have died. The cluster is bootstrapped automatically the first time the cluster is deployed. 
- Nodes that are no longer a part of the quorum will report `Non-Primary` when queried with `SHOW VARIABLES LIKE 'wsrep_cluster_status'`. See [Determining Cluster State](cluster-state.md) for more information.
- Cluster failure will occur if the cluster loses quorum (less than half of the nodes can communicate with each other). Once quorum is lost, the nodes will stop responding to write queries.  See [Cluster Behavior](cluster-behavior.md) for more details.
- Once lost, quorum can only be regained via bootstrapping. Even if monit restarts the node processes or the vms are recreated by bosh, the nodes will not become responsive until the cluster is manually repaired.

## Bootstrapping

1. SSH to each node in the cluster and shut down the mariadb process.

  ```sh
  $ monit stop mariadb_ctrl
  ```

  Re-bootstrapping the cluster will not be successful unless all other nodes have been shut down.

1. Choose a node to bootstrap.

    Find the node with the highest transaction sequence number (seqno):
    
    - If a node shutdown gracefully, the seqno should be in the galera state file. 
        
        ```sh
        $ cat /var/vcap/store/mysql/grastate.dat | grep 'seqno:'
        ```
        
    - If the node crashed or was killed, the seqno in the galera state file should be `-1`. In this case, the seqno may be recoverable from the database. The following command will cause the database to start up, log the recovered sequence number, and then exit. 
        
        ```sh
        $ /var/vcap/packages/mariadb/bin/mysqld --wsrep-recover
        ```
        
        Scan the error log for the recovered sequence number (the last number after the group id (uuid) is the recovered seqno):
        
        ```sh
        $ grep "Recovered position" /var/vcap/sys/log/mysql/mysql.err.log | tail -1
        150225 18:09:42 mysqld_safe WSREP: Recovered position e93955c7-b797-11e4-9faa-9a6f0b73eb46:15
        ```
        
        Note: The galera state file will still say `seqno: -1` afterward.
        
    - If the node never connected to the cluster before crashing, it may not even have a group id (uuid in grastate.dat). In this case there's nothing to recover. Unless all nodes crashed this way, don't choose this node for bootstrapping.

    Use the node with the highest `seqno` value as the new bootstrap node. If all nodes have the same `seqno`, you can choose any node as the new bootstrap node.

  **Important:** Only perform these bootstrap commands on the node with the highest `seqno`. Otherwise the node with the highest `seqno` will be unable to join the new cluster (unless its data is abandoned). Its mariadb process will exit with an error. See [cluster behavior](cluster-behavior.md) for more details on intentionally abandoning data.

1. On the new bootstrap node, update state file and restart the mariadb process:

  ```sh
  $ echo -n "NEEDS_BOOTSTRAP" > /var/vcap/store/mysql/state.txt
  $ monit start mariadb_ctrl
  ```

  You can check that the mariadb process has started successfully by running:

  ```sh
  $ watch monit summary
  ```

  It can take up to 10 minutes for monit to start the mariadb process.

1. Once the bootstrapped node is running, start the mariadb process on the remaining nodes via monit.

  ```sh
  $ monit start mariadb_ctrl
  ```

1. Verify that the new nodes have successfully joined the cluster. The following command should output the total number of nodes in the cluster:

  ```sh
  mysql> show status like 'wsrep_cluster_size';
  ```
