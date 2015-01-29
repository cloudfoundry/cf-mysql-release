# Bootstrapping a Galera Cluster

Bootstrapping is the process of (re)starting a Galera cluster.  

## When to Bootstrap

This is typically necessary in two scenarios:

1. No cluster exists yet and you wish to create one.
- A cluster loses quorum. Quorum is defined as greater than half the nodes in the last cluster with quorum. Quorum is lost when enough nodes die or become inaccessible such that no remaining subset of nodes has quorum. Once lost, quorum can only ever be regained via bootstrapping.

Bootstrapping from scenario #1 is automated during the initial deployment of cf-mysql-release. However, bootstrapping from scenario #2 requires manual intervention. The manual procedure for recovering from scenario #2 is documented below.

## Determining Cluster State

See [Determining Cluster State](cluster-state.md)

**Important:** You should only bootstrap the cluster if all nodes report a `Non-Primary` value for `wsrep_cluster_status`.

## Bootstrapping

If you have determined that it is necessary to bootstrap:

1. Find a new bootstrap node. Connect to each node and run the following mysql query:

    <pre class="terminal">
    mysql> show status like 'wsrep_last_committed';
    </pre>

    Use the node with the highest `wsrep_last_committed` value as the new bootstrap node.

- SSH to each node in the cluster and shut down the mariadb process:

    <pre class="terminal">
    $ monit stop mariadb_ctrl
    </pre>

- On the new bootstrap node, restart the mariadb process:

    <pre class="terminal">
    $ /var/vcap/packages/mariadb/bin/mysqld_safe --wsrep-new-cluster &
    </pre>


- In a new terminal, start the mariadb process on the remaining nodes one by one via monit.

    <pre class="terminal">
    $ monit start mariadb_ctrl
    </pre>

- Verify that the new nodes have successfully joined the cluster. The following command should output the total number of nodes in the cluster:

    <pre class="terminal">
    mysql> show status like 'wsrep_cluster_size';
    </pre>

- Because we started the bootstrap node without monit, we now need to restart the node using the normal monit script. On the bootstrap node, run:

    <pre class="terminal">
    $ /var/vcap/packages/mariadb/support-files/mysql.server stop
    $ monit start mariadb
    </pre>
