# Bootstrapping a Galera Cluster

Bootstrapping is the process of (re)starting a Galera cluster.  

## When to Bootstrap

This is typically necessary in two scenarios:

1. No cluster exists yet and you wish to create one.
- A cluster loses quorum. Quorum is defined as greater than half the nodes in the last cluster with quorum. Quorum is lost when enough nodes die or become inaccessible such that no remaining subset of nodes has quorum. Once lost, quorum can only ever be regained via bootstrapping.

Bootstrapping from scenario #1 is automated during the initial deployment of cf-mysql-release. However, bootstrapping from scenario #2 requires manual intervention. The manual procedure for recovering from scenario #2 is documented below.

## Determining Cluster State

See [Determining Cluster State](cluster-state.html)

**Important:** You should only bootstrap the cluster if all nodes report a `Non-Primary` value for `wsrep_cluster_status`.

## Bootstrapping

The start control script used by monit abstracts the complexity of bootstrapping a Galera cluster; this makes bootstrapping as simple as using monit to first start node 0, then the other nodes. As the cluster is currently configured to only accept connections on node 0, it is safe to assume that node 0 has the most recent data. For this reason, the monit control script is hardcoded to only bootstrap node 0.

If you have determined that it is necessary to bootstrap:

1. Shut down the mariadb process on each node in the cluster:

    <pre class="terminal">
    $ monit stop mariadb
    </pre>

- On Node 0, delete the file which records that the node has been bootstrapped:

    <pre class="terminal">
    $ rm /var/vcap/store/mysql/state.txt
    </pre>

- On Node 0, restart the mariadb process:

    <pre class="terminal">
    $ monit start mariadb
    </pre>

    The control script will start the node in bootstrap mode because the `state.txt` file will be absent.

- On Node 0, wait for the following to show that mariadb is running:

    <pre class="terminal">
    $ watch monit status
    </pre>

- Start the mariadb process on the remaining nodes one by one.

    <pre class="terminal">
    $ monit start mariadb
    </pre>
