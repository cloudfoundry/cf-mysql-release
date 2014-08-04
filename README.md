# Cloud Foundry MySQL Service

This project contains a BOSH release of a MySQL service for Cloud Foundry. It utilizes the [v2 broker API](http://docs.cloudfoundry.org/services/api.html).

## MySQL Service Components

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
 	   <td>MySQL 5.6 Server managed by the broker.  Database instances are hosted on this server.
 	   </td>
 	   <td> n/a </td>
 	 </tr>
</table>

## Installation

Prerequisites:

- The MySQL service requires a deployment of Cloud Foundry ([cf-release](https://github.com/cloudfoundry/cf-release)) and has been supported since [final release 155](https://github.com/cloudfoundry/cf-release/blob/master/releases/cf-155.yml) ([tag v155](https://github.com/cloudfoundry/cf-release/tree/v155)).
  - Dashboard SSO depends on having at least [version 169 of cf-release](https://github.com/cloudfoundry/cf-release/tree/v169). See the [Dashboard](#dashboard) section at the end of this doc for details.
- Installing the CF MySQL service requires BOSH.
- Instructions on installing BOSH as well as Cloud Foundry (runtime) are located in the [Cloud Foundry documentation](http://docs.cloudfoundry.org/).

Steps:

1. [Upload a supported stemcell](#upload_stemcell)
1. [Upload a release to the BOSH director](#upload_release)
1. [Create a CF MySQL deployment manifest](#create_manifest)
1. [Deploy the CF MySQL release with BOSH](#deploy_release)
1. [Register the service broker with Cloud Foundry](#register_broker)

After installation, the MySQL service should be shown when running `gcf marketplace`

### Upload Stemcell<a name="upload_stemcell"></a>

The latest final release, v10, expects Ubuntu Trusty go_agent 2652 by default. This final release may not work with newer stemcells as it has dependencies on libraries which have been removed from newer stemcells.

Master currently expects Ubuntu Trusty go_agent 2657 by default ([aws](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/cf-infrastructure-aws.yml#L14) | [vsphere](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/cf-infrastructure-vsphere.yml#L15)).

Stemcells can be downloaded from http://boshartifacts.cfapps.io/file_collections?type=stemcells.

### Upload Release<a name="upload_release"></a>

You can use a pre-built final release or build a release from HEAD. Final releases contain pre-compiled packages, making deployment much faster. However, these are created manually and infrequently. To be sure you're deploying the latest code, build a release yourself.

#### Upload a pre-built final BOSH release

1. Check out the tag for the desired version. This is necessary for generating a manifest that matches the code you're deploying.

  ```bash
  cd ~/workspace/cf-mysql-release
  ./update
  git checkout v10
  ```

1. Run the upload command, referencing one of the config files in the `releases` directory.

  ```bash
  bosh upload release releases/cf-mysql-10.yml
  ```

#### Create a BOSH Release from HEAD and Upload:

1. Build a BOSH development release from HEAD

  ```bash
  cd ~/workspace/cf-mysql-release
  ./update
  bosh create release
  ```

  When prompted to name the release, call it `cf-mysql`.

1. Upload the release to your bosh environment:

  ```bash
  bosh upload release
  ```

### Create a Manifest and Deploy<a name="create_manifest"></a>

#### BOSH-lite<a name="bosh-lite"></a>

1. Run the script [`bosh-lite/make_manifest`](bosh-lite/make_manifest) to generate your manifest for bosh-lite. This script uses a stub provided for you, `bosh-lite/stub.yml`. For a description of the parameters in this stub, see <a href="#manifest-stub-parameters">Manifest Stub Parameters</a> below.

    ```
    $ ./bosh-lite/make_manifest
    ```
    The manifest will be written to `bosh-lite/manifests/cf-riak-cs-manifest.yml`, which can be modified to change deployment settings.

1. The `make_manifest` script will set the deployment to `bosh-lite/manifests/cf-riak-cs-manifest.yml` for you, so to deploy you only need to run `bosh deploy`.

#### vSphere<a name="vsphere"></a>

1. Create a stub file called `cf-riak-cs-vsphere-stub.yml` that contains the properties in the example below. For a description of these and other manifest properties, see <a href="#manifest-stub-parameters">Manifest Stub Parameters</a> below.

    This stub differs from the bosh-lite stub in that it requires:

    * Username and password for admin user to support errands
    * Network settings, with 6 static IPs and 6+ dynamic IPs

  ```
  director_uuid: YOUR-DIRECTOR-GUID
  networks:
  - name: riak-cs-network
    subnets:
    - cloud_properties:
        name: YOUR-VSPHERE-NETWORK-NAME
      dns:
      - 8.8.8.8
      gateway: 10.0.0.1
      range: 10.0.0.0/24
      reserved:           # IPs that bosh should not use inside your subnet range
      - 10.0.0.2-10.0.0.99
      - 10.0.0.115-10.0.0.254
      static:
      - 10.0.0.100
      - 10.0.0.101
      - 10.0.0.102
      - 10.0.0.103
      - 10.0.0.104
      - 10.0.0.105
  properties:
    domain: YOUR-CF-SYSTEM-DOMAIN
    nats:
      machines:
      - 10.0.0.15   # IP of nats server
      user: NATS-USERNAME
      password: NATS-PASSWORD
      port: 4222
    cf:
      api_url: https://api.YOUR-CF-SYSTEM-DOMAIN
      apps_domain: YOUR-CF-APP-DOMAIN
      admin_username: CF-ADMIN-USERNAME
      admin_password: CF-ADMIN-PASSWORD
  ```

2. Generate the manifest: `./generate_deployment_manifest vsphere cf-riak-cs-vsphere-stub.yml > cf-riak-cs-vsphere.yml`
To tweak the deployment settings, you can modify the resulting file `cf-riak-cs-vsphere.yml`.

3. To deploy: `bosh deployment cf-riak-cs-vsphere.yml && bosh deploy`

#### AWS<a name="aws"></a>

1. Create a stub file called `cf-riak-cs-aws-stub.yml` that contains the parameters in the example below. For a description of these and other manifest properties, see <a href="#manifest-stub-parameters">Manifest Stub Parameters</a> below.

    This stub differs from the bosh-lite stub in that it requires:

    * Username and password for admin user to support errands
    * Network and resource pool settings

  ```
  director_uuid: YOUR-DIRECTOR-GUID
  networks:
  - name: riak-cs-network
    subnets:
    - name: riak-cs-subnet
      cloud_properties:
        subnet: YOUR-AWS-SERVICES-SUBNET-ID
  resource_pools:
  - name: riak-pool
    cloud_properties:
      availability_zone: YOUR-PRIMARY-AZ-NAME
  - name: broker-pool
    cloud_properties:
      availability_zone: YOUR-PRIMARY-AZ-NAME
  properties:
    domain: YOUR-CF-SYSTEM-DOMAIN
    nats:
      machines:
      - IP-OF-NATS-SERVER
      user: NATS-USERNAME
      password: NATS-PASSWORD
      port: 4222
    cf:
      api_url: https://api.YOUR-CF-SYSTEM-DOMAIN
      apps_domain: YOUR-CF-APP-DOMAIN
      admin_username: CF-ADMIN-USERNAME
      admin_password: CF-ADMIN-PASSWORD
  ```

1. Generate the manifest: `./generate_deployment_manifest aws cf-riak-cs-aws-stub.yml > cf-riak-cs-aws.yml`
To tweak the deployment settings, you can modify the resulting file `cf-riak-cs-aws.yml`.

1. To deploy: `bosh deployment cf-riak-cs-aws.yml && bosh deploy`

#### Deployment Manifest Stub Parameters<a name="stub-properties"></a>

### Deploy Using BOSH<a name="deploy_release"></a>

Set your deployment using the deployment manifest you generated above.

```bash
bosh deployment ~/workspace/deployments/mydevenv/cf-mysql-mydevenv.yml
bosh deploy
```

If you followed the instructions for bosh-lite above, your manifest is in the `cf-mysql-release/bosh-lite/manifests` directory. The make\_manifest\_spiff\_mysql script should have already set the deployment to the manifest, so you just have to run:

```bash
bosh deploy
```

### Register the Service Broker<a name="register_broker"></a>

#### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

```bash
bosh run errand broker-registrar
```

Note: the broker-registrar errand will fail if the broker has already been registered, and the broker name does not match the manifest property `jobs.broker-registrar.properties.broker.name`. Use the `cf rename-service-broker` CLI command to change the broker name to match the manifest property then this errand will succeed.

#### Manually

1. First register the broker using the `cf` CLI.  You must be logged in as an admin.

    ```bash
    cf create-service-broker p-mysql BROKER_USERNAME BROKER_PASSWORD URL
    ```

    `BROKER_USERNAME` and `BROKER_PASSWORD` are the credentials Cloud Foundry will use to authenticate when making API calls to the service broker. Use the values for manifest properties `jobs.cf-mysql-broker.properties.auth_username` and `jobs.cf-mysql-broker.properties.auth_password`.

    `URL` specifies where the Cloud Controller will access the MySQL broker. Use the value of the manifest property `jobs.cf-mysql-broker.properties.external_host`.

    For more information, see [Managing Service Brokers](http://docs.cloudfoundry.org/services/managing-service-brokers.html).

2. Then [make the service plan public](http://docs.cloudfoundry.org/services/managing-service-brokers.html#make-plans-public).

### Acceptance Tests<a name="acceptance_tests"></a>

To run the MySQL Release Acceptance tests, you will need:
- a running CF instance
- credentials for a CF Admin user
- a deployed MySQL Release with the broker registered and the plan made public
- an environment variable `$CONFIG` which points to a `.json` file that contains the application domain

#### BOSH errand

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

```bash
bosh run errand acceptance-tests
```

#### Manually

1. Install **Go** by following the directions found [here](http://golang.org/doc/install)
2. `cd` into `cf-mysql-release/test/acceptance-tests/`
3. Update `cf-mysql-release/test/acceptance-tests/integration_config.json`

    The following script will configure these prerequisites for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite)
installation. Replace credentials and URLs as appropriate for your environment.

```bash
#! /bin/bash

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
      "max_storage_mb": 10
    },
    {
      "plan_name": "1gb-dev",
      "max_storage_mb": 20
    }
  ],
  "skip_ssl_validation": true,
  "max_user_connections": 40
}
EOF
export CONFIG=$PWD/integration_config.json
```

    When `skip_ssl_validation: true`, commands run by the tests will accept self-signed certificates from Cloud Foundry. This option requires v6.0.2 or newer of the CLI.

4. Run  the tests

```bash
./bin/test
```

### De-register the Service Broker<a name="deregister_broker"></a>

The following commands are destructive and are intended to be run in conjuction with deleting your BOSH deployment.

#### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

This errand runs the two commands listed in the manual section below from a BOSH-deployed VM. This errand should be run before deleting your BOSH deployment. If you have already deleted your deployment follow the manual instructions below.

```bash
bosh run errand broker-deregistrar
```

#### Manually

Run the following:

```bash
cf purge-service-offering p-mysql
cf delete-service-broker p-mysql
```

### Dashboard <a name="dashboard"></a>

The service broker implements a user-facing UI which users can access via Single Sign-On (SSO) once authenticated with Cloud Foundry. SSO was implemented in build 169 of cf-release, so CF 169 is a minimum requirement for the SSO feature. If you encounter an error when you register the service broker, try removing the following lines from your manifest and redeploy.

        dashboard_client:
          id: p-mysql
          secret: yoursecret

Services wanting to implement such a UI and integrate with the Cloud Foundry Web UI should try something similar. Instructions to implement this feature can be found [here](http://docs.cloudfoundry.org/services/dashboard-sso.html).

The broker displays usage information on a per instance basis.

#### SSL

The dashboard URL defaults to using the `https` scheme. To override this, you can change `properties.ssl_enabled` to `false` in the `cf-mysql-broker` job.

Keep in mind that changing the `ssl_enabled` setting for an existing broker will not update previously advertised dashboard URLs.
Visiting the old URL may fail if you are using the [SSO integration](http://docs.cloudfoundry.org/services/dashboard-sso.html),
because the OAuth2 client registered with UAA will expect users to both come from and return to a URI using the scheme
implied by the `ssl_enabled` setting.

#### Implementation Details

1. Update the broker catalog with the dashboard client [properties](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/settings.yml#L26)
2. Implement oauth [workflow](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/config/initializers/omniauth.rb) with the [omniauth-uaa-oauth2 gem](https://github.com/cloudfoundry/omniauth-uaa-oauth2)
3. [Use](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/lib/uaa_session.rb) the [cf-uaa-lib gem](https://github.com/cloudfoundry/cf-uaa-lib) to get a valid access token and request permissions on the instance
4. Before showing the user the dashboard, [the broker checks](https://github.com/cloudfoundry/cf-mysql-broker/blob/master/app/controllers/manage/instances_controller.rb#L7) to see if the user is logged-in and has permissions to view the usage details of the instance.
