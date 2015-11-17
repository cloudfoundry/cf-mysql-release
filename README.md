# Cloud Foundry MySQL Service

### Table of contents

[Components](#components)

[Downloading a Stable Release](#downloading-a-stable-release)

[Development](#development)

[Release notes & known issues](#release-notes)

[Deploying](#deploying)

[Registering the Service Broker](#registering-broker)

[Security Groups](#security-groups)

[Smoke Tests and Acceptance Tests](#smoke-acceptance-tests)

[Deregistering the Service Broker](#deregistering-broker)

[Additional Configuration Options](#additional-configuration-options)

[CI](http://www.github.com/cloudfoundry-incubator/cf-mysql-ci)

<a name='components'></a>
## Components

A BOSH release of a MySQL database-as-a-service for Cloud Foundry using [MariaDB Galera Cluster](https://mariadb.com/kb/en/mariadb/documentation/replication-cluster-multi-master/galera/what-is-mariadb-galera-cluster/) and a [v2 Service Broker](http://docs.cloudfoundry.org/services/).

<table>
  <tr>
      <th>Component</th><th>Description</th>
  </tr>
  <tr>
    <td><a href="https://github.com/cloudfoundry/cf-mysql-broker">CF MySQL Broker</a></td>
    <td>Advertises the MySQL service and plans.  Creates and deletes MySQL databases and
    credentials (bindings) at the request of Cloud Foundry's Cloud Controller.
    </td>
   </tr>
   <tr>
     <td>MySQL Server</td>
     <td>MariaDB 10.0.17; database instances are hosted on the servers.</td>
   </tr>
      <tr>
     <td>Proxy</td>
     <td><a href="https://github.com/cloudfoundry-incubator/switchboard">Switchboard</a>; proxies to MySQL, severing connections on MySQL node failure.</td>
   </tr>
</table>

<a name='proxy'></a>
### Proxy

Traffic to the MySQL cluster is routed through one or more proxy nodes. The current proxy implementation is [Switchboard](https://github.com/cloudfoundry-incubator/switchboard). This proxy acts as an intermediary between the client and the MySQL server, providing failover between MySQL nodes. The number of nodes is configured by the proxy job instance count in the deployment manifest.

**NOTE:** If the number of proxy nodes is set to zero, apps will be bound to the IP address of the first MySQL node in the cluster. If that IP address should change for any reason (e.g. loss of a VM) or a proxy was subsequently added, one would need to re-bind all apps to the IP address of the new node.

For more details see the [proxy documentation](/docs/proxy.md).

<a name="dashboard"></a>
### Dashboard

A user-facing service dashboard is provided by the service broker that displays storage utilization information for each service instance.
The dashboard is accessible by users via Single Sign-On (SSO) once authenticated with Cloud Foundry.
The dashboard URL can be found by running `cf service MY_SERVICE_INSTANCE`.

Service authors interested in implementing a service dashboard accessible via SSO can follow documentation for [Dashboard SSO](http://docs.cloudfoundry.org/services/dashboard-sso.html).

#### Prerequisites

1. SSO is initiated when a user navigates to the URL found in the `dashboard_url` field. This value is returned to cloud controller by the broker in response to a provision request, and is exposed in the cloud controller API for the service instance. A users client must expose this field as a link, or it can be obtained via curl (`cf curl /v2/service_instances/:guid`) and copied into a browser.

2. SSO requires the following OAuth client to be configured in cf-release. This client is responsible for creating the OAuth client for the MySQL dashboard. Without this client configured in cf-release, the MySQL dashboard will not be accessible but the service will be otherwise functional. Registering the broker will display a warning to this effect.

    ```
    properties:
        uaa:
          clients:
            cc-service-dashboards:
              secret: cc-broker-secret
              scope: cloud_controller.write,openid,cloud_controller.read,cloud_controller_service_permissions.read
              authorities: clients.read,clients.write,clients.admin
              authorized-grant-types: client_credentials
    ```

3. SSO was implemented in v169 of cf-release; if you are on an older version of cf-release you'll encounter an error when you register the service broker. If upgrading cf-release is not an option, try removing the following lines from the cf-mysql-release manifest and redeploy.

    ```bash
    dashboard_client:
      id: p-mysql
      secret: yoursecret
    ```

#### Implementation Notes

The following links show how this release implements [Dashboard SSO](http://docs.cloudfoundry.org/services/dashboard-sso.html) integration.

1. Update the broker catalog with the dashboard client [properties](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/settings.yml#L26)
2. Implement oauth [workflow](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/initializers/omniauth.rb) with the [omniauth-uaa-oauth2 gem](https://github.com/cloudfoundry/omniauth-uaa-oauth2)
3. [Use](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/lib/uaa_session.rb) the [cf-uaa-lib gem](https://github.com/cloudfoundry/cf-uaa-lib) to get a valid access token and request permissions on the instance
4. Before showing the user the dashboard, [the broker checks](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/app/controllers/manage/instances_controller.rb#L7) to see if the user is logged-in and has permissions to view the usage details of the instance.

### Broker Configuration

#### Require HTTPS when visiting Dashboard

The dashboard URL defaults to using the `https` scheme. This means any requests using `http` will automatically be redirected to `https` instead.
To override this, you can change `jobs.cf-mysql-broker_z1.ssl_enabled` to `false`.

Keep in mind that changing the `ssl_enabled` setting for an existing broker will not update previously advertised dashboard URLs.
Visiting the old URL may fail if you are using the [SSO integration](http://docs.cloudfoundry.org/services/dashboard-sso.html),
because the OAuth2 client registered with UAA will expect users to both come from and return to a URI using the scheme
implied by the `ssl_enabled` setting.

Note:
If using `https`, the broker must be reached through an SSL termination proxy.
Connecting to the broker directly on `https` will result in a `port 443: Connection refused` error.

#### Trust Self-Signed SSL Certificates

By default, the broker will not trust a self-signed SSL certificate when communicating with cf-release.
To trust self-signed SSL certificates, you can change `jobs.cf-mysql-broker_z1.skip_ssl_validation` to `true`.

<a name='downloading-a-stable-release'></a>
## Downloading a Stable Release

Final releases are designed for public use, and are tagged with a version number of the form "v<N>".
The releases and corresponding release notes can be found on [github](https://github.com/cloudfoundry/cf-mysql-release/releases).

<a name='development'></a>
## Development

See our [contributing docs](CONTRIBUTING.md) for instructions on how to make a pull request.

This BOSH release doubles as a `$GOPATH`. It will automatically be set up for
you if you have [direnv](http://direnv.net) installed.

    # fetch release repo
    mkdir -p ~/workspace
    cd ~/workspace
    git clone https://github.com/cloudfoundry/cf-mysql-release.git
    cd cf-mysql-release/

    # switch to develop branch (not master!)
    git checkout develop

    # automate $GOPATH and $PATH setup
    direnv allow

    # initialize and sync submodules
    ./update

If you do not wish to use direnv, you can simply `source` the `.envrc` file in the root
of the release repo.  You may manually need to update your `$GOPATH` and `$PATH` variables
as you switch in and out of the directory.

<a name='release-notes'></a>
## Release Notes & Known Issues

For release notes and known issues, see [the release wiki](https://github.com/cloudfoundry/cf-mysql-release/wiki/).

<a name='deploying'></a>
## Deploying

### Prerequisites

- A deployment of [BOSH](https://github.com/cloudfoundry/bosh)
- A deployment of [Cloud Foundry](https://github.com/cloudfoundry/cf-release), [final release 193](https://github.com/cloudfoundry/cf-release/tree/v193) or greater
- Instructions for installing BOSH and Cloud Foundry can be found at http://docs.cloudfoundry.org/.

### Overview

1. [Upload Stemcell](#upload_stemcell)
1. [Upload Release](#upload_release)
1. [Create Infrastructure](#create_infrastructure)
1. [Deployment Components](#deployment_components)
1. [Create Manifest and Deploy](#create_manifest)

After installation, the MySQL service will be visible in the Services Marketplace; using the [CLI](https://github.com/cloudfoundry/cli), run `cf marketplace`.

<a name="upload_stemcell"></a>
### Upload Stemcell

The latest final release expects the Ubuntu Trusty (14.04) go_agent stemcell version [2859](https://github.com/cloudfoundry/bosh/blob/master/CHANGELOG.md#2859) by default. Older stemcells are not recommended. Stemcells can be downloaded from http://bosh.io/stemcells; choose the appropriate stemcell for your infrastructure ([vsphere esxi](https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/vsphere/bosh-stemcell-2859-vsphere-esxi-ubuntu-trusty-go_agent.tgz) or [aws hvm](https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-2859-aws-xen-hvm-ubuntu-trusty-go_agent.tgz)).

<a name="upload_release"></a>
### Upload Release

You can use a pre-built final release or build a dev release from any of the branches described in <a href="#branches">Getting the Code</a>.

Final releases are stable releases created periodically for completed features. They also contain pre-compiled packages, which makes deployment much faster. To deploy the latest final release, simply check out the **master** branch. This will contain the latest final release and accompanying materials to generate a manifest. If you would like to deploy an earlier final release, use `git checkout <tag>` to obtain both the release and corresponding manifest generation materials. It's important that the manifest generation materials are consistent with the release.

If you'd like to deploy the latest code, build a release yourself from the **develop** branch.

#### Upload a pre-built final BOSH release

Run the upload command, referencing the latest config file in the `releases` directory.

  ```
  $ cd ~/workspace/cf-mysql-release
  $ git checkout master
  $ ./update
  $ bosh upload release releases/cf-mysql-<N>.yml
  ```

If deploying an **older** final release than the latest, check out the tag for the desired version; this is necessary for generating a manifest that matches the code you're deploying.

  ```
  $ cd ~/workspace/cf-mysql-release
  $ git checkout v<N>
  $ ./update
  $ bosh upload release releases/cf-mysql-<N>.yml
  ```

#### Create and upload a BOSH Release:

1. Checkout one of the branches described in <a href="#branches">Getting the Code</a>. Build a BOSH development release.

  ```
  $ cd ~/workspace/cf-mysql-release
  $ git checkout release-candidate
  $ ./update
  $ bosh create release
  ```

  When prompted to name the release, call it `cf-mysql`.

1. Upload the release to your bosh environment:

  ```
  $ bosh upload release
  ```

<a name="create_infrastructure"></a>
### Create Infrastructure

Note: No infrastructure changes are required to deploy to bosh-lite

#### Define subnets

Prior to deployment, the operator should define three subnets via their infrastructure provider.
The MySQL release is designed to be deployed across three subnets to ensure availability in the event of a subnet failure.  During installation, a fourth subnet is required for compilation vms.
The [sample_aws_stub.yml](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/sample_stubs/sample_aws_stub.yml) demonstrates how these subnets can be configured on AWS across multiple availability zones.

#### Create load balancer

In order to route requests to both proxies, the operator should create a load balancer.
Manifest changes required to configure a load balancer can be found in the
[proxy](https://github.com/cloudfoundry/cf-mysql-release/blob/master/docs/proxy.md#configuring-load-balancer) documentation.
Once a load balancer is configured, the brokers will hand out the address of the load balancer rather than the IP of the first proxy.
Currently, load balancing requests across both proxies can increase the possibility of deadlocks. See the [routing](https://github.com/cloudfoundry/cf-mysql-release/blob/master/docs/proxy.md#consistent-routing) documentation for more information.
To avoid this problem, configure the load balancer to route requests to the second proxy only in the event of a failure.

<a name="deployment_components"></a>
### Deployment Components

#### Database nodes

There are three mysql jobs (mysql\_z1, mysql\_z2, mysql\_z3) which should be deployed with one instance each.
Each of these instances will reside in separate subnets as described in the previous section.
The number of mysql nodes should always be odd, with a minimum count of three, to avoid [split-brain](http://en.wikipedia.org/wiki/Split-brain\_\(computing\)).
When the failed node comes back online, it will automatically rejoin the cluster and sync data from one of the healthy nodes. Note: Due to our bootstrapping procedure, if you are bringing up a cluster for the first time, there must be a database node in the first subnet.

#### Proxy nodes

There are two proxy jobs (proxy\_z1, proxy\_z2), which should be deployed with one instance each to different subnets.
The second proxy is intended to be used in a failover capacity. In the event the first proxy fails, the second proxy will still be able to route requests to the mysql nodes.

#### Broker nodes

There are also two broker jobs (cf-mysql-broker\_z1, cf-mysql-broker\_z2) which should be deployed with one instance each to different subnets.
The brokers each register a route with the router, which load balances requests across the brokers.

<a name="create_manifest"></a>
### Create Manifest and Deploy

<a name="bosh-lite"></a>
#### BOSH-lite

1. Generate the manifest using a bosh-lite specific script and a stub provided for you, `bosh-lite/cf-mysql-stub-spiff.yml`.

    ```
    $ ./bosh-lite/make_manifest
    ```
    The resulting file, `bosh-lite/manifests/cf-mysql-manifest.yml` is your deployment manifest. To modify the deployment configuration, you can edit the stub and regenerate the manifest or edit the manifest directly.

1. The `make_manifest` script will set the deployment to `bosh-lite/manifests/cf-mysql-manifest.yml` for you, so to deploy you only need to run:
  ```
  $ bosh deploy
  ```

<a name="vsphere"></a>
#### vSphere

1. Create a stub file called `cf-mysql-vsphere-stub.yml` by copying and modifying the [sample_vsphere_stub.yml](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/sample_stubs/sample_vsphere_stub.yml)  in `templates/sample_stubs`. The `sample_plans_stub.yml` can also be copied if values need changing.

2. Generate the manifest:
  ```
  $ ./generate_deployment_manifest \
    vsphere \
    plans_stub.yml \
    cf-mysql-vsphere-stub.yml > cf-mysql-vsphere.yml
  ```
  The resulting file, `cf-mysql-vsphere.yml` is your deployment manifest. To modify the deployment configuration, you can edit the stub and regenerate the manifest or edit the manifest directly.

3. To deploy:
  ```
  $ bosh deployment cf-mysql-vsphere.yml && bosh deploy
  ```

<a name="aws"></a>
#### AWS

1. Create a stub file called `cf-mysql-aws-stub.yml` by copying and modifying the [sample_aws_stub.yml](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/sample_stubs/sample_aws_stub.yml) in `templates/sample_stubs`. The `sample_plans_stub.yml` can also be copied if values need changing.

1. Generate the manifest:
  ```
  $ ./generate_deployment_manifest \
    aws \
    plans_stub.yml \
    cf-mysql-aws-stub.yml > cf-mysql-aws.yml
  ```
  The resulting file, `cf-mysql-aws.yml` is your deployment manifest. To modify the deployment configuration, you can edit the stub and regenerate the manifest or edit the manifest directly.

1. To deploy:
  ```
  $ bosh deployment cf-mysql-aws.yml && bosh deploy
  ```

<a name="manifest-properties"></a>
#### Deployment Manifest Properties

Manifest properties are described in the `spec` file for each job; see [jobs](jobs).

You can find your `director_uuid` by running `bosh status`.

The MariaDB cluster nodes are configured by default with 100GB of persistent disk. This can be configured in your stub or manifest using `disk_pools.mysql-persistent-disk.disk_size`, however your deployment will fail if this is less than 3GB; we recommend allocating 10GB at a minimum.

<a name="registering-broker"></a>
## Registering the Service Broker

### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

```
$ bosh run errand broker-registrar
```

Note: the broker-registrar errand will fail if the broker has already been registered, and the broker name does not match the manifest property `jobs.broker-registrar.properties.broker.name`. Use the `cf rename-service-broker` CLI command to change the broker name to match the manifest property then this errand will succeed.

### Manually

1. First register the broker using the `cf` CLI.  You must be logged in as an admin.

    ```
    $ cf create-service-broker p-mysql BROKER_USERNAME BROKER_PASSWORD URL
    ```

    `BROKER_USERNAME` and `BROKER_PASSWORD` are the credentials Cloud Foundry will use to authenticate when making API calls to the service broker. Use the values for manifest properties `jobs.cf-mysql-broker_z1.properties.auth_username` and `jobs.cf-mysql-broker_z1.properties.auth_password`.

    `URL` specifies where the Cloud Controller will access the MySQL broker. Use the value of the manifest property `jobs.cf-mysql-broker_z1.properties.external_host`. By default, this value is set to `p-mysql.<properties.domain>` (in spiff: `"p-mysql." .properties.domain`).

    For more information, see [Managing Service Brokers](http://docs.cloudfoundry.org/services/managing-service-brokers.html).

2. Then [make the service plan public](http://docs.cloudfoundry.org/services/managing-service-brokers.html#make-plans-public).

## Security Groups

Note: adding additional security groups for cf-mysql is not required on bosh-lites running cf-release [v212](https://github.com/cloudfoundry/cf-release/blob/v212/bosh-lite/cf-stub-spiff.yml#L47) or later.

Since [cf-release](https://github.com/cloudfoundry/cf-release) v175, applications by default cannot to connect to IP addresses on the private network. This prevents applications from connecting to the MySQL service. To enable access to the service, create a new security group for the IP configured in your manifest for the property `jobs.cf-mysql-broker_z1.mysql_node.host`.

1. Add the rule to a file in the following json format; multiple rules are supported.

  ```
  [
		{
			"destination": "10.10.163.1-10.10.163.255",
			"protocol": "all"
		},
		{
			"destination": "10.10.164.1-10.10.164.255",
			"protocol": "all"
		},
		{
			"destination": "10.10.165.1-10.10.165.255",
			"protocol": "all"
		}
	]
  ```
- Create a security group from the rule file.
  ```shell
  $ cf create-security-group p-mysql rule.json
  ```

- Enable the rule for all apps
  ```
  $ cf bind-running-security-group p-mysql
  ```

Security group changes are only applied to new application containers; existing apps must be restarted.

<a name="smoke-acceptance-tests"></a>
## Smoke Tests and Acceptance Tests

The smoke tests are a subset of the acceptance tests, useful for verifying a deployment. The acceptance tests are for developers to validate changes to the MySQL Release. These tests can be run manually or from a BOSH errand. For details on running these tests manually, see [Acceptance Tests](docs/acceptance-tests.md).

The MySQL Release contains an "acceptance-tests" job which is deployed as a BOSH errand. The errand can then be run to verify the deployment. A deployment manifest [generated with the provided spiff templates](#create_manifest) will include this job. The errand can be configured to run either the smoke tests (default) or the acceptance tests.

<a name="smoke_tests"></a>
### Running Smoke Tests via BOSH errand

To run the MySQL Release Smoke tests you will need:

- a running CF instance
- credentials for a CF Admin user
- a deployed MySQL Release with the broker registered and the plan made public

Run the smoke tests via bosh errand as follows:

```
$ bosh run errand acceptance-tests
```

Modifying values under `jobs.acceptance-tests.properties` may be required. Configuration options can be found in the [job spec](jobs/acceptance-tests/spec).

<a name="deregistering-broker"></a>
## De-registering the Service Broker

The following commands are destructive and are intended to be run in conjuction with deleting your BOSH deployment.

### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

This errand runs the two commands listed in the manual section below from a BOSH-deployed VM.

```
$ bosh run errand broker-deregistrar
```

### Manually

Run the following:

```
$ cf purge-service-offering p-mysql
$ cf delete-service-broker p-mysql
```

<a name="deployment-resources"></a>
## Deployment Resources

The service is configured to have a small footprint out of the box. These resources are sufficient for development, but may be insufficient for production workloads. If the service appears to be performing poorly, redeploying with increased resources may improve performance. See [deployment resources](docs/deployment-resources.md) for further details.

<a name="additional-configuration-options"></a>
## Additional Configuration Options

### Updating Service Plans

Updating the service instances is supported; see [Service plans and instances](docs/service-plans-instances.md) for details.

### Pre-seeding Databases

Normally databases are created via the `cf create-service` command, and
a MySQL user is created and given access to that database when an app is bound to that service instance.
However, it is sometimes useful to have databases and users already available when the service is deployed,
without having to run `cf create-service` or bind an app.
To specify any preseeded databases, add the following to the deployment manifest:

```
jobs:
- name: mysql_z1
  properties:
    seeded_databases:
    - name: db1
      username: user1
      password: pw1
    - name: db2
      username: user2
      password: pw2
```

Note: If all you need is a database deployment, it is possible to deploy this
release with zero broker instances and completely remove any dependencies on Cloud Foundry.
See the [proxy](jobs/proxy/spec) and [acceptance-tests](jobs/acceptance-tests/spec) spec files for standalone configuration options.

### Configuring how long the startup script waits for the database to come online

On larger databases, the default database startup timeout may be too low.
This would result in the job reporting as failing, while MySQL continues to bootstrap in the background (see [Known Issues > Long SST Transfers](docs/Known-Issues.md#long-sst-transfers)).
To increase the duration that the startup script waits for MySQL to start, add the following to your deployment stub:

```yaml
jobs:
- name: mysql_z1
  properties:
    database_startup_timeout: 360
```

Note: This is independent of the overall BOSH timeout which is also configurable in the manifest. The BOSH timeout should always be higher than the database startup timeout:

```yaml
update:
  canary_watch_time: 30000-600000
  update_watch_time: 30000-600000
```
