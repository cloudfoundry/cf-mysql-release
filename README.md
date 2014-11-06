# Cloud Foundry MySQL Service

A BOSH release of a MySQL database-as-a-service for Cloud Foundry using [MariaDB Galera Cluster](https://mariadb.com/kb/en/mariadb/documentation/replication-cluster-multi-master/galera/what-is-mariadb-galera-cluster/) and a [v2 Service Broker](http://docs.cloudfoundry.org/services/).

<table>
  <tr>
     	<th>Component</th><th>Description</th><th>Build Status</th>
 	</tr>
 	<tr>
 	  <td>CF MySQL Broker</td>
 	  <td>Advertises the MySQL service and plans.  Creates and deletes MySQL databases and
 	  credentials (bindings) at the request of Cloud Foundry's Cloud Controller.
 	  </td>
 	  <td><a href="https://travis-ci.org/cloudfoundry/cf-mysql-broker"><img src="https://travis-ci.org/cloudfoundry/cf-mysql-broker.png" alt="Build Status"></a></td>
 	 </tr>
 	 <tr>
 	   <td>MySQL Server</td>
 	   <td>MariaDB 10.0.12; database instances are hosted on the servers.
 	   </td>
 	   <td> n/a </td>
 	 </tr>
</table>

## Release Notes & Known Issues

For release notes and known issues, see [the release wiki](https://github.com/cloudfoundry/cf-mysql-release/wiki/).

## Getting the code

There are multiple branches, representing code in different stages of development.
* master is the latest and greatest (untested) code - the bleeding-edge. Use at your own risk.
* release-candidate has passed both automated and manual acceptance. This branch should be suitable for most use cases.
* Final releases are tagged with the version number e.g. v12.

## Deployment

### Prerequisites

- A deployment of [BOSH](https://github.com/cloudfoundry/bosh)
- A deployment of [Cloud Foundry](https://github.com/cloudfoundry/cf-release), [final release 169](https://github.com/cloudfoundry/cf-release/tree/v169) or greater
- Instructions for installing BOSH and Cloud Foundry can be found at http://docs.cloudfoundry.org/.

### Overview

1. [Upload Stemcell](#upload_stemcell)
1. [Upload Release](#upload_release)
1. [Create Manifest and Deploy](#create_manifest)
1. [Register the Service Broker](#register_broker)

After installation, the MySQL service will be visible in the Services Marketplace; using [the CLI](https://github.com/cloudfoundry/cli) run `cf marketplace`.

### Upload Stemcell<a name="upload_stemcell"></a>

The latest final release, v12, expects the Ubuntu Trusty (14.04) go_agent stemcell version 2682 by default (Note: there are important security fixes not in this stemcell.  We recommend using stmcell 2719.3 or later). This release may not work with newer stemcells it may have dependencies on libraries which have been removed from newer stemcells. Stemcells can be downloaded from http://boshartifacts.cfapps.io/file_collections?type=stemcells; choose the ubuntu trusty go_agent stemcell 2682 for your infrastructure ([vsphere esxi](https://s3.amazonaws.com/bosh-jenkins-artifacts/bosh-stemcell/vsphere/bosh-stemcell-2682-vsphere-esxi-ubuntu-trusty-go_agent.tgz) or [aws xen](https://s3.amazonaws.com/bosh-jenkins-artifacts/bosh-stemcell/aws/bosh-stemcell-2682-aws-xen-ubuntu-trusty-go_agent.tgz)).

Master expects Ubuntu Trusty go_agent by default, but the version changes frequently.

Stemcells can be downloaded from http://boshartifacts.cfapps.io/file_collections?type=stemcells.

### Upload Release<a name="upload_release"></a>

You can use a pre-built final release or build a release from HEAD. Final releases contain pre-compiled packages, making deployment much faster. However, these are created manually and infrequently. To be sure you're deploying the latest code, build a release yourself.

#### Upload a pre-built final BOSH release

1. Check out the tag for the desired version. This is necessary for generating a manifest that matches the code you're deploying.

  ```
  $ cd ~/workspace/cf-mysql-release
  $ ./update
  $ git checkout v12
  $ git submodule update --recursive
  ```

1. Run the upload command, referencing one of the config files in the `releases` directory.

  ```
  $ bosh upload release releases/cf-mysql-12.yml
  ```

#### Create a BOSH Release from HEAD and Upload:

1. Build a BOSH development release from HEAD

  ```
  $ cd ~/workspace/cf-mysql-release
  $ ./update
  $ bosh create release
  ```

  When prompted to name the release, call it `cf-mysql`.

1. Upload the release to your bosh environment:

  ```
  $ bosh upload release
  ```

### Create Manifest and Deploy<a name="create_manifest"></a>

#### BOSH-lite<a name="bosh-lite"></a>

1. Generate the manifest using a bosh-lite specific script and a stub provided for you, `bosh-lite/cf-mysql-stub-spiff.yml`.

    ```
    $ ./bosh-lite/make_manifest_spiff_mysql
    ```
    The resulting file, `bosh-lite/manifests/cf-mysql-manifest.yml` is your deployment manifest. To modify the deployment configuration, you can edit the stub and regenerate the manifest or edit the manifest directly.

1. The `make_manifest_spiff_mysql` script will set the deployment to `bosh-lite/manifests/cf-mysql-manifest.yml` for you, so to deploy you only need to run:
  ```
  $ bosh deploy
  ```

#### vSphere<a name="vsphere"></a>

1. Create a stub file called `cf-mysql-vsphere-stub.yml` by copying and modifying the [sample_vsphere_stub.yml](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/sample_stubs/sample_vsphere_stub.yml)  in `templates/sample_stubs`.

2. Generate the manifest:
  ```
  $ ./generate_deployment_manifest vsphere cf-mysql-vsphere-stub.yml > cf-mysql-vsphere.yml
  ```
  The resulting file, `cf-mysql-vsphere.yml` is your deployment manifest. To modify the deployment configuration, you can edit the stub and regenerate the manifest or edit the manifest directly.

3. To deploy:
  ```
  $ bosh deployment cf-mysql-vsphere.yml && bosh deploy
  ```

#### AWS<a name="aws"></a>

1. Create a stub file called `cf-mysql-aws-stub.yml` by copying and modifying the [sample_aws_stub.yml](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/sample_stubs/sample_aws_stub.yml) in `templates/sample_stubs`.

1. Generate the manifest:
  ```
  $ ./generate_deployment_manifest aws cf-mysql-aws-stub.yml > cf-mysql-aws.yml
  ```
  The resulting file, `cf-mysql-aws.yml` is your deployment manifest. To modify the deployment configuration, you can edit the stub and regenerate the manifest or edit the manifest directly.

1. To deploy:
  ```
  $ bosh deployment cf-mysql-aws.yml && bosh deploy
  ```

#### Deployment Manifest Properties<a name="manifest-properties"></a>

Manifest properties are described in the `spec` file for each job; see [jobs](jobs).

You can find your director_uuid by running `bosh status`.

The MariaDB cluster nodes are configured by default with 100GB of persistent disk. This can be configured in your stub or manifest using `jobs.mysql.persistent_disk`, however your deployment will fail if this is less than 3GB; we recommend allocating 10GB at a minimum.

## Register the Service Broker<a name="register_broker"></a>

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

    `BROKER_USERNAME` and `BROKER_PASSWORD` are the credentials Cloud Foundry will use to authenticate when making API calls to the service broker. Use the values for manifest properties `jobs.cf-mysql-broker.properties.auth_username` and `jobs.cf-mysql-broker.properties.auth_password`.

    `URL` specifies where the Cloud Controller will access the MySQL broker. Use the value of the manifest property `jobs.cf-mysql-broker.properties.external_host`.

    For more information, see [Managing Service Brokers](http://docs.cloudfoundry.org/services/managing-service-brokers.html).

2. Then [make the service plan public](http://docs.cloudfoundry.org/services/managing-service-brokers.html#make-plans-public).

## Acceptance Tests<a name="acceptance_tests"></a>

To run the MySQL Release Acceptance tests, you will need:
- a running CF instance
- credentials for a CF Admin user
- a deployed MySQL Release with the broker registered and the plan made public
- an environment variable `$CONFIG` which points to a `.json` file that contains the application domain

### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

The following properties must be included in the manifest (most will be there by default):
- cf.api_url:
- cf.admin_username:
- cf.admin_password:
- cf.apps_domain:
- cf.skip_ssl_validation:
- broker.host:
- service.name:
- service.plans:

The service.plans array must include the following properties for each plan:
- plan_name:
- max_storage_mb:

To customize the following values add them to the manifest:
- mysql.max_user_connections: (default: 40)

To run the errand:

```
$ bosh run errand acceptance-tests
```

### Manually

1. Install **Go** by following the directions found [here](http://golang.org/doc/install)
2. `cd` into `cf-mysql-release/src/acceptance-tests/`
3. Update `cf-mysql-release/src/acceptance-tests/integration_config.json`

    The following commands provide a shortcut to configuring `integration_config.json` with values for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite)
deployment. Copy and paste this into your terminal, then open the resulting `integration_config.json` in an editor to replace values as appropriate for your environment.

  ```bash
  cat > integration_config.json <<EOF
  {
    "api_url": "http://api.10.244.0.34.xip.io",
    "apps_domain": "10.244.0.34.xip.io",
    "admin_user": "admin",
    "admin_password": "admin",
    "broker_host": "p-mysql.10.244.0.34.xip.io",
    "service_name": "p-mysql",
    "plans" : [
      {
        "plan_name": "100mb-dev",
        "max_user_connections": 20,
        "max_storage_mb": 10
      },
      {
        "plan_name": "1gb-dev",
        "max_user_connections": 40,
        "max_storage_mb": 20
      }
    ],
    "skip_ssl_validation": true,
    "timeout_scale": 1.0,
    "include_dashboard_tests": false,
    "include_failover_tests": false
  }
  EOF
  export CONFIG=$PWD/integration_config.json
  ```

  When `skip_ssl_validation: true`, commands run by the tests will accept self-signed certificates from Cloud Foundry. This option requires v6.0.2 or newer of the CLI.

  All timeouts in the test suite can be scaled proportionally by changing the `timeout_scale` factor.

4. Run  the tests

  ```
  $ ./bin/test
  ```

## Security Groups

Since [cf-release](https://github.com/cloudfoundry/cf-release) v175, applications by default cannot to connect to IP addresses on the private network. This prevents applications from connecting to the MySQL service. To enable access to the service, create a new security group for the IP configured in your manifest for the property `jobs.mysql_broker.mysql_node.host`.

1. Add the rule to a file in the following json format; multiple rules are supported.

  ```
  [
      {
        "destination": "10.244.1.18",
        "protocol": "all"
      }
  ]
  ```
- Create a security group from the rule file.
  <pre class="terminal">
  $ cf create-security-group p-mysql rule.json
  </pre>
- Enable the rule for all apps
  <pre class="terminal">
  $ cf bind-running-security-group p-mysql
  </pre>

Changes are only applied to new application containers; in order for an existing app to receive security group changes it must be restarted.

## De-register the Service Broker<a name="deregister_broker"></a>

The following commands are destructive and are intended to be run in conjuction with deleting your BOSH deployment.

### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

This errand runs the two commands listed in the manual section below from a BOSH-deployed VM. This errand should be run before deleting your BOSH deployment. If you have already deleted your deployment follow the manual instructions below.

```
$ bosh run errand broker-deregistrar
```

### Manually

Run the following:

```
$ cf purge-service-offering p-mysql
$ cf delete-service-broker p-mysql
```

## Dashboard <a name="dashboard"></a>

The service broker implements a user-facing UI which users can access via Single Sign-On (SSO) once authenticated with Cloud Foundry. SSO was implemented in build 169 of cf-release, so CF 169 is a minimum requirement for the SSO feature. If you encounter an error when you register the service broker, try removing the following lines from your manifest and redeploy.

  ```bash
  dashboard_client:
    id: p-mysql
    secret: yoursecret
  ```

Services wanting to implement such a UI and integrate with the Cloud Foundry Web UI should try something similar. Instructions to implement this feature can be found [here](http://docs.cloudfoundry.org/services/dashboard-sso.html).

The broker displays usage information on a per instance basis.

### SSL

The dashboard URL defaults to using the `https` scheme. To override this, you can change `properties.ssl_enabled` to `false` in the `cf-mysql-broker` job.

Keep in mind that changing the `ssl_enabled` setting for an existing broker will not update previously advertised dashboard URLs.
Visiting the old URL may fail if you are using the [SSO integration](http://docs.cloudfoundry.org/services/dashboard-sso.html),
because the OAuth2 client registered with UAA will expect users to both come from and return to a URI using the scheme
implied by the `ssl_enabled` setting.

### Implementation Details

1. Update the broker catalog with the dashboard client [properties](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/settings.yml#L26)
2. Implement oauth [workflow](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/initializers/omniauth.rb) with the [omniauth-uaa-oauth2 gem](https://github.com/cloudfoundry/omniauth-uaa-oauth2)
3. [Use](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/lib/uaa_session.rb) the [cf-uaa-lib gem](https://github.com/cloudfoundry/cf-uaa-lib) to get a valid access token and request permissions on the instance
4. Before showing the user the dashboard, [the broker checks](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/app/controllers/manage/instances_controller.rb#L7) to see if the user is logged-in and has permissions to view the usage details of the instance.

## HAProxy

Traffic to the MySQL cluster is routed through an HAProxy node. The intention is to use this proxy for fail-over and load balancing traffic.

The HAProxy node runs a thin web server to display traffic stats. This dashboard can be reached at `haproxy-1.p-mysql.<system domain>`. Log in as user `admin` with the password configured by the `haproxy_stats_password` property in the deployment manifest.

