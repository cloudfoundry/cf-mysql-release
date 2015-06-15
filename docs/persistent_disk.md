# Recovering from issues with persistent disk(s)

Problems with persistent disks typically are caused by infrastructure failure (ex: an administrator disables/deletes/detaches a disk) or hardware failure (ex: a disk physically breaks). The MariaDB nodes store their data on the persistent disk, so it is neccessary to either re-attach or re-create the disk before restarting the node.

In order to determine what recovery is required, first determine the state of the Galera cluster; see [Determining Cluster State](cluster-state.md).

### If cluster has failed

If the cluster has lost quorum (less than half of the nodes can communicate), the rest of the nodes should automatically shut down. 

1. Any corrupted or lost persistent disks must be repaired or replaced. 
1. The cluster must be bootstrapped to restart it. See the [bootstrapping docs](bootstrapping.md) for more details.

### If the cluster is still running

If clustering has quorum (at least half of the nodes can communicate), an administrator only needs to recover the failing nodes. 

1. Any corrupted or lost persistent disks must be repaired or replaced. 
2. Recovered nodes must be restarted.
3. The recovered node should automatically rejoin the cluster and be updated with the current data.

See below for recovery instructions.

## Disk Recovery

The process to recover depends on the the nature of the disk failure.

### When persistent disk is detached and can be re-attached

When the disk is detached, monit considers the process stopped and BOSH will consider the job as failing. However currently BOSH cloud check will not recognize the disk is unattached without a little kick.

1. Attempt to recreate the broken node using BOSH. This will alert BOSH to the missing disk.
  <pre class="terminal">
  $ bosh recreate mysql_AZ INDEX
  </pre>
  You should see an error that looks like this:
  <pre class="terminal">
  Error 100: Disk (YOUR_DISK_ID) is not attached to VM (vm-fc4ab74e-61ed-4d12-aa93-a1bbb389723f)
  </pre>
  If recreate fails with the following error, wait until monit times out and all jobs are stopped then try again.
  <pre class="terminal">
  Failed updating job mysql_AZ: mysql_AZ/INDEX (canary) (00:00:22): Action Failed get_task: Task 8ace0778-c5aa-4a2f-55a0-42443452adb1 result: Stopping Monitored Services: Stopping service gra-log-purger-executable: Stopping Monit service gra-log-purger-executable: Request failed with 503 Service Unavailable:
  </pre>
- Use BOSH cloud check to reattach the disk.
  <pre class="terminal">
  $ bosh cck
  </pre>
  When prompted, choose `3. Reattach disk and reboot instance`; this should succeed. As BOSH recreate failed after stopping the jobs, BOSH believes the jobs should stay stopped on reboot.
- Upon restarting the node again, it will join the cluster.
  <pre class="terminal">
  $ bosh restart mysql_AZ INDEX
  </pre>

### When persistent disk is lost and needs to be re-created

1. ssh into cf-mysql-broker instances and stop the processes. This will prevent creation and deletion of instances when we attempt to recreate the broken node to get its disk id.
  <pre class="terminal">
  $ sudo monit stop all
  </pre>
- ssh into proxy instances and stop the processes. This will prevent data from changing on the broken node when we attempt to recreate it to get its disk id.
  <pre class="terminal">
  $ sudo monit stop all
  </pre>
- Attempt to recreate the node using BOSH in order to obtain the disk id.
  <pre class="terminal">
  $ bosh recreate mysql_AZ INDEX
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
- Through the infrastructure interface (e.g., vCenter client or AWS console), power off and delete the VM corresponding to `mysql_AZ/INDEX`.
- Use BOSH cloud check to remove reference to the VM.
  <pre class="terminal">
  $ bosh cck
  </pre>
  When prompted, choose `3. Delete VM reference (DANGEROUS!)`.
- Edit the deployment manifest and reduce the number of instances of both cf-mysql-broker and proxy to 0 (also remove the static ip for proxy). This will prevent new connections from being made to the node after it is recreated but before it joins the cluster.
- Deploy the release to recreate the node and remove the broker and proxy.
  <pre class="terminal">
  $ bosh deploy
  </pre>
- ssh into any one of the nodes and verify that all nodes have joined the cluster; for instructions, see [Determining Cluster State](cluster-state.md).
- Only after all nodes have joined the cluster should you edit the deployment manifest, setting the number of instances for cf-mysql-broker and proxy back to 1 and restoring the static ip for proxy. Then deploy the release.
  <pre class="terminal">
  $ bosh deploy
  </pre>

## Simulating disk failure

To simulate a disk issues in vCenter, follow these steps:

1. Log in to vCenter client
- Locate the vm that correlates to `mysql_z1/0`. You can do this most easily by running `bosh vms` and identifying the IP address of the VM. The vCenter client will match VMs to IP addresses.
- Navigate to the VM detail view, then click Edit.
- Locate the entry for the persistent disk and delete it. You will have the option to detach the disk or delete it from the datastore.

Note:
- It may take a few minutes for MariaDB to fail. Attempting to write to a database will trigger this immediately.
- `bosh recreate mysql_z1 0` should reveal that the disk has been detached. `bosh cck` will not report any issues until an operator has first attempted to recreate the VM with `bosh recreate`.
