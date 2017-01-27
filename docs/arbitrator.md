# Arbitrator node

As part of CF MySQL v26, we provide an
[arbitrator node](http://galeracluster.com/documentation-webpages/arbitrator.html)
as a replacement for one of the MySQL nodes. The arbitrator is a Galera node
which does not participate in database transactions or data replication, but
just votes in order to maintain quorum. This enables the user to save on resource
cost (since the arbitrator is pretty lightweight compared to a normal database node),
while still avoiding split-brain conditions which can happen when one of the
database nodes is unreachable.

In other words, we are going from a 3-node configuration to a
2-node-plus-arbitrator configuration as part of v26. In a typical bosh
deployment of CF MySQL, the node is called `arbitrator_z3` and replaces the
earlier `mysql_z3` node.

### Deploying 2-node-plus-arbitrator configuration as fresh deploy (or upgrading from an existing 3-node deployment)

For a normal deployment of CF MySQL v26, simply follow the steps in the
[README](https://github.com/cloudfoundry/cf-mysql-release/blob/develop/README.md#create-manifest-and-deploy).

### Rolling from 2-node-plus-arbitrator back to 3-node deployment

If you wish to go back to a 3-node deployment from a 2-node-plus-arbitrator
deployment (and have not retained your old cf-mysql manifest), follow the steps
below to perform the downgrade:

1. Generate a 3-node manifest, for example:
```
./scripts/generate-deployment-manifest \
  -c <YOUR_CONFIG_REPO>/cf.yml \
  -p <YOUR_CONFIG_REPO>/cf-mysql/property-overrides.yml \
  -i <YOUR_CONFIG_REPO>/cf-mysql/iaas-settings.yml \
  -n manifest-generation/examples/no-arbitrator/instance-count-overrides.yml
  > [manifest export path].yml
  ```
1. `bosh deployment [manifest export path].yml`
1. `bosh deploy`

#### IMPORTANT:

If you are performing rolling upgrades of MySQL on a bosh-lite environment,
you must note that there is a known issue (https://github.com/cloudfoundry/bosh-lite/issues/193).
This means that by default, bosh-lite masquerades source IPs of inter-container packets.
These masqueraded IPs can confuse Galera cluster, especially when nodes
shutdown and restart (as during a rolling upgrade). The fix for this is to
turn off the NAT rule that changes the source address of inter-container traffic,
by doing a `vagrant ssh` to your bosh-lite box, and then running the following as root user:

```
iptables -I POSTROUTING -t nat --source 10.244.0.0/16 --destination 10.244.0.0/16 --jump ACCEPT
```

Without doing this, the above two rolling deploy cases **will not work**. The command can also be added to the Vagrantfile so that it remains part of the configuration when your bosh-lite box is destroyed and recreated.
