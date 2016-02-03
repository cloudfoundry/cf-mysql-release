# Arbitrator node

As part of CF MySQL v26, we provide an [arbitrator node](http://galeracluster.com/documentation-webpages/arbitrator.html) as a replacement for one of the MySQL nodes. The arbitrator is a Galera node which does not participate in database transactions or data replication, but just votes in order to maintain quorum. This enables the user to save on resource cost (since the arbitrator is pretty lightweight compared to a normal database node), while still avoiding split-brain conditions which can happen when one of the database nodes is unreachable.

In other words, we are going from a 3-node configuration to a 2-node-plus-arbitrator configuration as part of v26. In a typical bosh deployment of CF MySQL, the node is called `arbitrator_z3` and replaces the earlier `mysql_z3` node.

### Deploying 2-node-plus-arbitrator configuration as fresh deploy

For a fresh deployment of CF MySQL v26, simply follow the steps in the [README](https://github.com/cloudfoundry/cf-mysql-release/blob/develop/README.md#create-manifest-and-deploy).

### Upgrading to 2-node-plus-arbitrator from already existing 3-node deployment

If you already have a 3-node deployment of CF MySQL (e.g. v25), follow the below steps to upgrade to the new configuration as a rolling deploy (the commands below apply to testing on a bosh-lite deployment. For other environments, the same steps are to be followed, except that stubs are different, as explained in the README linked above.).

1. Generate a 3 node + arbitrator manifest:
    <pre>./scripts/generate-deployment-manifest -c /tmp/bosh-lite-cf-manifest.yml -p manifest-generation/bosh-lite-stubs/property-overrides.yml -i manifest-generation/bosh-lite-stubs/iaas-settings.yml -n manifest-generation/examples/upgrade-to-arbitrator/deploy-arbitrator/instance-count-overrides.yml > [manifest export path].yml</pre>
1. `bosh deployment [manifest export path].yml`
1. `bosh deploy`
1.  Upon successful deployment, generate the 2 node + arbitrator manifest:
    <pre>./scripts/generate-deployment-manifest -c /tmp/bosh-lite-cf-manifest.yml -p manifest-generation/bosh-lite-stubs/property-overrides.yml -i manifest-generation/bosh-lite-stubs/iaas-settings.yml -n manifest-generation/examples/upgrade-to-arbitrator/remove-mysql-node/instance-count-overrides.yml > [manifest export path].yml</pre>
1. `bosh deployment [manifest export path].yml`
1. `bosh deploy`

### Rolling from 2-node-plus-arbitrator back to 3-node deployment

If you wish to go back to a 3-node deployment from a 2-node-plus-arbitrator deployment, follow the steps below to perform the downgrade:

1. Generate a 3-node + arbitrator manifest:
    <pre>./scripts/generate-deployment-manifest -c /tmp/bosh-lite-cf-manifest.yml -p manifest-generation/bosh-lite-stubs/property-overrides.yml -i manifest-generation/bosh-lite-stubs/iaas-settings.yml -n manifest-generation/examples/upgrade-to-arbitrator/deploy-arbitrator/instance-count-overrides.yml > [manifest export path].yml</pre>
1. `bosh deployment [manifest export path].yml`
1. `bosh deploy`
1.  Upon successful deployment, generate the 3 node manifest:
    <pre>./scripts/generate-deployment-manifest -c /tmp/bosh-lite-cf-manifest.yml -p manifest-generation/bosh-lite-stubs/property-overrides.yml -i manifest-generation/bosh-lite-stubs/iaas-settings.yml -n manifest-generation/examples/no-arbitrator/instance-count-overrides.yml > [manifest export path].yml</pre>
1. `bosh deployment [manifest export path].yml`
1. `bosh deploy`
