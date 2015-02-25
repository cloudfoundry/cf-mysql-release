# Bootstrapping a Galera Cluster

Bootstrapping is the process of (re)starting a Galera cluster.  

## When to Bootstrap

- Manual bootstrapping should only be required if all nodes have died. The cluster is bootstrapped automatically the first time the cluster is deployed. 
- Nodes that are no longer a part of the quorum will report `Non-Primary` when queried with `SHOW VARIABLES LIKE 'wsrep_cluster_status'`. See [Determining Cluster State](cluster-state.md) for more information.
- Cluster failure will occur if the cluster loses quorum (less than half of the nodes can communicate with each other). Once quorum is lost, the nodes will stop responding to write queries.  See [Cluster Behavior](cluster-behavior.md) for more details.
- Once lost, quorum can only be regained via bootstrapping. Even if monit restarts the node processes or the vms are recreated by bosh, the nodes will not become responsive until the cluster is manually repaired.

## Bootstrapping

1. SSH to each node in the cluster and shut down the mariadb process.

  <pre class="terminal">
  $ monit stop mariadb_ctrl
  </pre>

  Re-bootstrapping the cluster will not be successful unless all other nodes have been shut down.

1. Find a new bootstrap node. Connect to each node and run the following command to find each node's `seqno`:

  <pre class="terminal">
  $ cat /var/vcap/store/mysql/grastate.dat | grep 'seqno:'
  </pre>

  Use the node with the highest `seqno` value as the new bootstrap node. If all nodes have the same `seqno`, you can choose any node as the new bootstrap node.

  **Important:** Only perform these bootstrap commands on the node with the highest `seqno`. Otherwise the node with the highest `seqno` will be unable to join the new cluster, as its mariadb process will not start.

1. On the new bootstrap node, update state file and restart the mariadb process:

  <pre class="terminal">
  $ echo -n "NEEDS_BOOTSTRAP" > /var/vcap/store/mysql/state.txt
  $ monit start mariadb_ctrl
  </pre>

  You can check that the mariadb process has started successfully by running:

  <pre class="terminal">
  $ monit summary
  </pre>

  It can take up to 10 minutes for monit to start the mariadb process.

1. In a new terminal, start the mariadb process on the remaining nodes one by one via monit.

  <pre class="terminal">
  $ monit start mariadb_ctrl
  </pre>

1. Verify that the new nodes have successfully joined the cluster. The following command should output the total number of nodes in the cluster:

  <pre class="terminal">
  mysql> show status like 'wsrep_cluster_size';
  </pre>
