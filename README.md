# Cloud Foundry MySQL Service

A BOSH release of a MySQL database-as-a-service for Cloud Foundry using [MariaDB Galera Cluster](https://mariadb.com/kb/en/mariadb/documentation/replication-cluster-multi-master/galera/what-is-mariadb-galera-cluster/) and a [v2 Service Broker](http://docs.cloudfoundry.org/services/).

<table>
  <tr>
      <th>Component</th><th>Description</th><th>Build Status</th>
  </tr>
  <tr>
    <td><a href="https://github.com/cloudfoundry/cf-mysql-broker">CF MySQL Broker</a></td>
    <td>Advertises the MySQL service and plans.  Creates and deletes MySQL databases and
    credentials (bindings) at the request of Cloud Foundry's Cloud Controller.
    </td>
    <td><a href="https://travis-ci.org/cloudfoundry/cf-mysql-broker"><img src="https://travis-ci.org/cloudfoundry/cf-mysql-broker.svg" alt="Build Status"></a></td>
   </tr>
   <tr>
     <td>MySQL Server</td>
     <td>MariaDB 10.0.16; database instances are hosted on the servers.</td>
     <td> n/a </td>
   </tr>
      <tr>
     <td>Proxy</td>
     <td><a href="https://github.com/cloudfoundry-incubator/switchboard">Switchboard</a>; proxies to MySQL, severing connections on MySQL node failure.</td>
     <td><a href="https://travis-ci.org/cloudfoundry-incubator/switchboard"><img src="https://travis-ci.org/cloudfoundry-incubator/switchboard.svg" alt="Build Status"></a></td>
   </tr>
</table>


<a name='branches'></a>
## Getting the code

Final releases are designed for public use, and are tagged with a version number of the form "v<N>".

The [**develop**](https://github.com/cloudfoundry/cf-mysql-release/tree/develop) branch is where we do active development. Although we endeavor to keep the [**develop**](https://github.com/cloudfoundry/cf-mysql-release/tree/develop) branch stable, we do not guarantee that any given commit will deploy cleanly.

The [**release-candidate**](https://github.com/cloudfoundry/cf-mysql-release/tree/release-candidate) branch has passed all of our unit, integration, smoke, & acceptance tests, but has not been used in a final release yet. This branch should be fairly stable.

The [**master**](https://github.com/cloudfoundry/cf-mysql-release/tree/master) branch points to the most recent stable final release.

At semi-regular intervals a final release is created from the [**release-candidate**](https://github.com/cloudfoundry/cf-mysql-release/tree/release-candidate) branch. This final release is tagged and pushed to the [**master**](https://github.com/cloudfoundry/cf-mysql-release/tree/master) branch.

Pushing to any branch other than [**develop**](https://github.com/cloudfoundry/cf-mysql-release/tree/develop) will create problems for the CI pipeline, which relies on fast forward merges. To recover from this condition follow the instructions [here](https://github.com/cloudfoundry/cf-release/blob/master/docs/fix_commit_to_master.md).

## Development

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

## Release Notes & Known Issues

For release notes and known issues, see [the release wiki](https://github.com/cloudfoundry/cf-mysql-release/wiki/).

## Deployment

### Prerequisites

- A deployment of [BOSH](https://github.com/cloudfoundry/bosh)
- A deployment of [Cloud Foundry](https://github.com/cloudfoundry/cf-release), [final release 193](https://github.com/cloudfoundry/cf-release/tree/v193) or greater
- Instructions for installing BOSH and Cloud Foundry can be found at http://docs.cloudfoundry.org/.

### Overview

1. [Upload Stemcell](#upload_stemcell)
1. [Upload Release](#upload_release)
1. [Create Manifest and Deploy](#create_manifest)
1. [Register the Service Broker](#register_broker)

After installation, the MySQL service will be visible in the Services Marketplace; using the [CLI](https://github.com/cloudfoundry/cli), run `cf marketplace`.

<a name="upload_stemcell"></a>
### Upload Stemcell

The latest final release expects the Ubuntu Trusty (14.04) go_agent stemcell version 2831 by default. Older stemcells are not recommended. Stemcells can be downloaded from http://bosh.io/stemcells; choose the appropriate stemcell for your infrastructure ([vsphere esxi](https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-2831-aws-xen-hvm-ubuntu-trusty-go_agent.tgz) or [aws hvm](https://d26ekeud912fhb.cloudfront.net/bosh-stemcell/aws/light-bosh-stemcell-2831-aws-xen-hvm-ubuntu-trusty-go_agent.tgz)).

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

<a name="aws"></a>
#### AWS

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

<a name="manifest-properties"></a>
#### Deployment Manifest Properties

Manifest properties are described in the `spec` file for each job; see [jobs](jobs).

You can find your director_uuid by running `bosh status`.

The MariaDB cluster nodes are configured by default with 100GB of persistent disk. This can be configured in your stub or manifest using `jobs.mysql.persistent_disk`, however your deployment will fail if this is less than 3GB; we recommend allocating 10GB at a minimum.

<a name="register_broker"></a>
## Register the Service Broker

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

<a name="tests"></a>
## Smoke Tests & Acceptance Tests

The smoke tests are a subset of the acceptance tests, useful for verifying a deployment. The acceptance tests are for developers to validate changes to the MySQL Release. These tests can be run manually or from a BOSH errand. For details on running these tests manually, see [Acceptance Tests](docs/acceptance-tests.md).

The MySQL Release contains an "acceptance-tests" job which is deployed as a BOSH errand. The errand can then be run to verify the deployment. A deployment manifest [generated with the provided spiff templates](#create_manifest) will include this job. The errand can be configured to run either the smoke tests (default) or the acceptance tests.

<a name="smoke_tests"></a>
### Running Smoke Tests via BOSH errand

To run the MySQL Release Smoke tests you will need:

- a running CF instance
- credentials for a CF Admin user
- a deployed MySQL Release with the broker registered and the plan made public

The following properties must be included in the deployment manifest under the `acceptance-tests` job (most will be there by default):

- `cf.api_url`
- `cf.admin_username`
- `cf.admin_password`
- `cf.apps_domain`
- `cf.skip_ssl_validation`
- `broker.host`
- `service.name`
- `service.plans`

The `service.plans` array must include the following properties for each plan:

- `plan_name`
- `max_storage_mb`

The following property is optional:

- `mysql.max_user_connections` (default: 40)

To run the smoke tests via bosh errand:

```
$ bosh run errand acceptance-tests
```


<a name="deregister_broker"></a>
## De-register the Service Broker

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

<a name="dashboard"></a>
## Dashboard

A user-facing service dashboard is provided by the service broker that displays storage utilization information for each service instance. The dashboard is accessible by users via Single Sign-On (SSO) once authenticated with Cloud Foundry.

Service authors interested in implementing a service dashboard accessible via SSO can follow documentation for [Dashboard SSO](http://docs.cloudfoundry.org/services/dashboard-sso.html).

### Prerequisites

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

3. SSO was implemented in v169 of cf-release; if you are on an older version of cf-release you'll encounter an error when you register the service broker. If upgradiing cf-release is not an option, try removing the following lines from the cf-mysql-release manifest and redeploy.

    ```bash
    dashboard_client:
      id: p-mysql
      secret: yoursecret
    ```

### SSL

The dashboard URL defaults to using the `https` scheme. To override this, you can change `properties.ssl_enabled` to `false` in the `cf-mysql-broker` job.

Keep in mind that changing the `ssl_enabled` setting for an existing broker will not update previously advertised dashboard URLs.
Visiting the old URL may fail if you are using the [SSO integration](http://docs.cloudfoundry.org/services/dashboard-sso.html),
because the OAuth2 client registered with UAA will expect users to both come from and return to a URI using the scheme
implied by the `ssl_enabled` setting.

### Implementation Notes

The following links show how this release implements [Dashboard SSO](http://docs.cloudfoundry.org/services/dashboard-sso.html) integration.

1. Update the broker catalog with the dashboard client [properties](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/settings.yml#L26)
2. Implement oauth [workflow](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/initializers/omniauth.rb) with the [omniauth-uaa-oauth2 gem](https://github.com/cloudfoundry/omniauth-uaa-oauth2)
3. [Use](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/lib/uaa_session.rb) the [cf-uaa-lib gem](https://github.com/cloudfoundry/cf-uaa-lib) to get a valid access token and request permissions on the instance
4. Before showing the user the dashboard, [the broker checks](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/app/controllers/manage/instances_controller.rb#L7) to see if the user is logged-in and has permissions to view the usage details of the instance.

## Proxy

More extensive proxy documentation can be found [here](https://github.com/cloudfoundry/cf-mysql-release/docs/proxy.md)


Traffic to the MySQL cluster is routed through one or more proxy nodes. The current proxy implementation is [Switchboard](https://github.com/cloudfoundry-incubator/switchboard). This proxy acts as an intermediary between the client and the MySQL server - providing failover between MySQL nodes. The number of nodes is configured by the job instance count in the deployment manifest.
