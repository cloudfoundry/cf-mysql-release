# Bootstrapping a Galera Cluster

Bootstrapping is the process of (re)starting a Galera cluster.

## When to Bootstrap

Bootstrapping is only required when the cluster has lost quorum.

Quorum is lost when less than half of the nodes can communicate with each other (for longer than the configured grace period). In Galera terminology, if a node can communicate with the rest of the cluster, its DB is in a good state, and it reports itself as ```synced```.

If quorum has *not* been lost, individual unhealthy nodes should automatically rejoin the cluster once repaired (error resolved, node restarted, or connectivity restored).

#### Symptoms of Lost Quorum

- [All nodes appear "Unhealthy" on the proxy dashboard.](quorum-lost.png)
- All responsive nodes report the value of `wsrep_cluster_status` as `non-Primary`.

    ```sh
    mysql> SHOW STATUS LIKE 'wsrep_cluster_status';
    +----------------------+-------------+
    | Variable_name        | Value       |
    +----------------------+-------------+
    | wsrep_cluster_status | non-Primary |
    +----------------------+-------------+
    ```

- All responsive nodes respond with `ERROR 1047` when queried with most statement types.

    ```sh
    mysql> select * from mysql.user;
    ERROR 1047 (08S01) at line 1: WSREP has not yet prepared node for application use
    ```

See [Cluster Behavior](cluster-behavior.md) for more details about determining cluster state.

## Auto-bootstrap errand

As part of cf-mysql-release v25, we provide an auto-bootstrap feature which runs as a BOSH errand. The errand evaluates if quorum has been lost on a cluster, and if so bootstraps the cluster. Before running the errand, one should ensure that there are no network partitions. Once network partitions have been resolved, the cluster is in a state where the errand can be run.

#### How to run

Run `bosh run errand bootstrap` from the terminal. When done, this should successfully bootstrap the cluster, and all jobs should report as `running`. Note that:

If the cluster was already healthy to begin with (i.e. quorum was never lost), the errand will error out saying `bootstrap is not required`.

If one or more nodes are not reachable (i.e. the VM exists but in an unknown state), it will error out saying `Error: could not reach node`. In this situation, follow the steps below:

1. `bosh -n stop mysql_z1 && bosh -n stop mysql_z2 && bosh -n stop <arbitrator|mysql>_z3`
1. `bosh edit deployment`
1. Set `update.canaries` to 0, `update.max_in_flight` to 3, and `update.serial` to false.
1. `bosh deploy`
  - Note, if you get a 503 error (like `Sending stop request to monit: Request failed, response: Response{ StatusCode: 503, Status: '503 Service Unavailable' }`), it means that monit is still trying to stop the vms. Please wait a few minutes and try this step again.
1. `bosh -n start mysql_z1 ; bosh -n start mysql_z2 ; bosh -n start <arbitrator|mysql>_z3`
  - This will throw several errors, but it ensures that all the jobs are present on the VM.
1. `bosh instances` to verify that all jobs report as failing.
1. Try running the errand again using `bosh -n run errand bootstrap` as above.
  - Once the errand succeeds, the cluster is synced, although some jobs might still report as failing.
1. `bosh edit deployment`
1. Set `update.canaries` to 1, `update.max_in_flight` to 1, and `update.serial` to true.
1. Verify that deployment succeeds and all jobs are healthy. A healthy deployment should look like this:

```
$ bosh vms cf-mysql'
Acting as user 'admin' on deployment 'cf-mysql' on 'Bosh Lite Director'
| mysql_z1/0           | running | mysql_z1           | 10.244.7.2   |
| mysql_z2/0           | running | mysql_z2           | 10.244.8.2   |
| arbitrator_z3/0      | running | arbitrator_z3      | 10.244.9.6   |
...
```

If these steps did not work for you, please refer to the [Manual Bootstrap Process](#manual-bootstrap-process) below.

## How it works

The bootstrap errand simply automates the steps in the manual bootstrapping process documented below. It finds the node with the highest transaction sequence number, and asks it to start up by itself (i.e. in bootstrap mode), then asks the remaining nodes to join the cluster.

The sequence number of a stopped node can be retained by either reading the node's state file under `/var/vcap/store/mysql/grastate.dat`, or by running a mysqld command with a WSREP flag, like `mysqld --wsrep-recover`.

## Manual Bootstrap Process

The following steps are prone to user-error and can result in lost data if followed incorrectly.
Please follow the [Auto-bootstrap](#auto-bootstrap-errand) instructions above first, and only resort to the manual process if the errand fails to repair the cluster.

1. SSH to each node in the cluster and, as root, shut down the mariadb process.

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
  mysql> SHOW STATUS LIKE 'wsrep_cluster_size';
  ```
