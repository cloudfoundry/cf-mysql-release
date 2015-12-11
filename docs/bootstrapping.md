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

Run ```bosh run errand bootstrap``` from the terminal. When done, this should successfully bootstrap the cluster, and all jobs should report as ```running```. Note that:

1. If the cluster was already healthy to begin with (i.e. quorum was never lost), the errand will error out saying ```bootstrap is not required```. 
1. If one or more nodes are not reachable (i.e. the VM exists but in an unknown state), it will error out saying ```nodes are not reachable```. In this situation, follow the steps below:
	1. ```bosh -n stop mysql_z1 && bosh -n stop mysql_z2 && bosh -n stop mysql_z3```
	1. ```bosh edit deployment```
	1. Set ```update.canaries``` to 0, ```update.max_in_flight``` to 3, and ```update.serial``` to false.
	1. ```bosh deploy```
	1. ```bosh -n start mysql_z1 || bosh -n start mysql_z2 || bosh -n start mysql_z3``` (This will throw several errors, but it ensures that all the jobs are present on the VM)
   1. ```bosh instances``` to verify that all jobs report as failing.
   1. Try running the errand again using ```bosh run errand bootstrap``` as above.
   1. Once the errand succeeds, the cluster is synced, although some jobs might still report as failing.
   1. ```bosh edit deployment```
   1. Set ```update.canaries``` to 1, ```update.max_in_flight``` to 1, and ```update.serial``` to true.
   1. Verify that deployment succeeds and all jobs are healthy. 
   
## How it works

The bootstrap errand simply automates the steps in the manual bootstrapping process documented in previous releases of cf-mysql. It finds the node with the highest transaction sequence number, and asks it to start up by itself (i.e. in bootstrap mode), then asks the remaining nodes to join the cluster.

The sequence number of a stopped node can be retained by either reading the node's state file under ```/var/vcap/store/mysql/grastate.dat```, or by running a mysqld command with a WSREP flag, like ```mysqld --wsrep-recover```. 
