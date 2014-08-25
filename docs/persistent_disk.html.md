# Recovering from issues with persistent disk

This section explains how to recover from problems with a node's persistent disk. These problems typically happen when there is an infrastructure failure
(for example, an administrator disables a disk), or hardware failure (for example, a disk physically breaks). The Mariadb nodes store their data on the
persistent disk, so it's important to be able to reliably re-attach the persistent disk (via BOSH) or to re-create it and immediately populate it with its
original data (via BOSH and Galera).

## Simulating disk failure

To simulate a disk issues in vCenter, follow these steps:

1. Log in to vCenter client
- Locate the vm that correlates to `mysql/0`. You can do this most easily by running `bosh vms` and identifying the IP address of the VM. The vCenter client will match VMs to IP addresses.
- Navigate to the VM detail view, then click Edit.
- Locate the entry for the persistent disk and delete it. You will have the option to detach the disk or delete it from the datastore.

Note:
- It may take a few minutes for MariaDB to fail. Attempting to write to a database will trigger this immediately.
- `bosh recreate mysql 0` should reveal that the disk has been detached. `bosh cck` will not report any issues until an operator has first attempted to recreate the VM with `bosh recreate`.

## Recovery

The first thing to is determine the state of the Galera cluster; see [Determining Cluster State](cluster-state.html.md).

If clustering is intact (no network partitions) and most of the nodes are functioning normally, an administrator only needs to recover the persistent disk for the node with disk issues. When the node with the detached or lost disk is recovered, it should successfully rejoin the cluster and Galera's replication should bring the node up-to-date. See [When all nodes are still in the primary component](#cluster-intact) below for recovery instructions.

On the other hand, if the cluster has lost quorum or most of the nodes are not running, the administrator must recreate/restart nodes in a particular order to preserve the cluster's data. This ensures that the node(s) that have the most up-to-date data don't try to join a cluster without that data. See [When cluster has lost quorum](#quorum-lost) below for recovery instructions.

<a id="cluster-intact"/>
### When all nodes are still in the primary component

This only requires that the operator recreate the one node with disk issues. The process to recover depends on the the nature of the disk failure.

<a id="detachedDisk"/>
#### When persistent disk is detached and can be re-attached

When the disk is detached, monit considers the process stopped and BOSH will consider the job as failing. However currently BOSH cloud check will not recognize the disk is unattached without a little kick.

1. Attempt to recreate Node 0 using BOSH. This will alert BOSH to the missing disk.
  <pre class="terminal">
  $ bosh recreate mysql 0
  </pre>
  You should see an error that looks like this:
  <pre class="terminal">
  Error 100: Disk (YOUR_DISK_ID) is not attached to VM (vm-fc4ab74e-61ed-4d12-aa93-a1bbb389723f)
  </pre>
  If recreate fails with the following error, wait until monit times out and all jobs are stopped then try again.
  <pre class="terminal">
  Failed updating job mysql: mysql/0 (canary) (00:00:22): Action Failed get_task: Task 8ace0778-c5aa-4a2f-55a0-42443452adb1 result: Stopping Monitored Services: Stopping service gra-log-purger-executable: Stopping Monit service gra-log-purger-executable: Request failed with 503 Service Unavailable:
  </pre>
- Use BOSH cloud check to reattach the disk.
  <pre class="terminal">
  $ bosh cck
  </pre>
  When prompted, choose `3. Reattach disk and reboot instance`; this should succeed. As BOSH recreate failed after stopping the jobs, BOSH believes the jobs should stay stopped on reboot.
- Upon restarting the node again, it will join the cluster.
  <pre class="terminal">
  $ bosh restart mysql 0
  </pre>

<a id="lostDisk"/>
#### When persistent disk is lost and needs to be re-created

##### If node 0 has lost its persistent disk:
1. ssh into cf-mysql-broker and stop the processes. This will prevent creation and deletion of instances when we attempt to recreate Node 0 to get its disk id.
  <pre class="terminal">
  $ sudo monit stop all
  </pre>
- ssh into haproxy and stop the processes. This will prevent data from changing on Node 0 when we attempt to recreate id to get its disk id.
  <pre class="terminal">
  $ sudo monit stop all
  </pre>
- Attempt to recreate Node 0 using BOSH in order to obtain the disk id.
  <pre class="terminal">
  $ bosh recreate mysql 0
  </pre>
  You should see an error that looks like this:
  <pre class="terminal">
  Error 100: Disk (YOUR_DISK_ID) is not attached to VM (vm-fc4ab74e-61ed-4d12-aa93-a1bbb389723f)
  </pre>

  Make a note of the value of `YOUR_DISK_ID`; you'll use it in the next step.

  If you attempt to use BOSH cloud check to reattach the disk and you see the following error, then you know the disk is lost.
  <pre class="terminal">
  $ bosh cck
  ...
  Failed: File []/vmfs/volumes/volume-id/datastore-id/disk-id.vmdk was not found
  </pre>
- Connect to the BOSH director postgres database and remove references to the disk.
  On a microbosh the database is local.
  <pre class="terminal">
  $ /var/vcap/packages/postgres/bin/psql -U postgres --password bosh
  psql (9.0.3)
  Type "help" for help.

  bosh=> DELETE FROM persistent_disks WHERE disk_cid = 'YOUR_DISK_ID';
  DELETE 1
  bosh=> DELETE FROM vsphere_disk WHERE uuid = 'YOUR_DISK_ID';
  DELETE 1
  </pre>

  This will delete the reference to the lost disk, and cause BOSH to recreate a fresh disk for this VM on the next deploy.
- Through the infrastructure interface (e.g., vCenter client or AWS console), power off and delete the VM corresponding to `mysql/0`.
- Use BOSH cloud check to remove reference to the VM.
  <pre class="terminal">
  $ bosh cck
  </pre>
  When prompted, choose `3. Delete VM reference (DANGEROUS!)`.
- Edit the deployment manifest and reduce the number of instances of both cf-mysql-broker and haproxy to 0 (also remove the static ip for haproxy). This will prevent new connections from being made to Node 0 after it is recreated but before it joins the cluster.
- Deploy the release to recreate Node 0 and remove the broker and haproxy.
  <pre class="terminal">
  $ bosh deploy
  </pre>
- ssh into any one of the nodes and verify that all nodes have joined the cluster; for instructions, see [Determining Cluster State](cluster-state.html.md).
- Only after all nodes have joined the cluster should you edit the deployment manifest, setting the number of instances for cf-mysql-broker and haproxy back to 1 and restoring the static ip for haproxy. Then deploy the release.
  <pre class="terminal">
  $ bosh deploy
  </pre>

##### If nodes 1 or 2 have lost their persistent disk:

Since all connections are routed to Node 0, we don't need to worry about new data being written to these nodes while they are not a member of the cluster.

1. Follow steps 3, 4, 5, and 6 above for the node that has lost its disk.
- Recreate the node. BOSH will create a new VM and attach a new disk. The node will join the cluster on startup.
  <pre class="terminal">
  $ bosh recreate mysql 1
  </pre>

<a id="quorum-lost"/>
### When cluster has lost quorum

1. ssh into all nodes and use monit to stop the MariaDB process.
  <pre class="terminal">
  $ sudo monit stop mariadb
  </pre>
- For each node with detached or lost disk:
  - Follow the instructions above for [When all nodes are still in the primary component](#cluster-intact) to reattach or recreating the disk.
  - If the disk for Node 0 has been recreated, the MariaDB process will bootstrap
  and write its state.txt file to persistent disk. All future attempts to restart this node will cause the node to JOIN rather than BOOTSTRAP.
  - ssh into the node and wait for `watch monit status` to show that the `mariadb` process is `running`
  - Once the process is running, stop it with `sudo monit stop mariadb`
- Choose the node with the data you want to restart your cluster with. We will call this the BOOTSTRAP node.
ssh into the BOOTSTRAP node and start it in bootstrap mode manually.
  <pre class="terminal">
  $ /var/vcap/packages/mariadb/support-files/mysql.server bootstrap
  </pre>
- Restart the `mariadb` process on all nodes except for the BOOTSTRAP node:
  <pre class="terminal">
  $ sudo monit start mariadb
  </pre>
- Wait for all remaining nodes to join and sync with the cluster.
- Finally, ssh into the BOOTSTRAP node, stop MariaDB manually, and start it with monit:
  <pre class="terminal">
  $ /var/vcap/packages/mariadb/support-files/mysql.server stop
  $ monit start mariadb
  </pre>
- Now the BOOTSTRAP node will rejoin the cluster under monit's supervision.
