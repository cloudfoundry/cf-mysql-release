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

The latest final release, v8, [expects Ubuntu Lucid ruby_agent 2366 by default](https://github.com/cloudfoundry/cf-mysql-release/blob/v8/templates/cf-mysql-template.yml#L41). This final release may not work with newer stemcells as it has dependencies on libraries which have been removed from newer stemcells.

Master currently expects Ubuntu Trusty go_agent 2657 by default ([aws](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/cf-infrastructure-aws.yml#L31) | [vsphere](https://github.com/cloudfoundry/cf-mysql-release/blob/master/templates/cf-infrastructure-vsphere.yml#L15)).

Stemcells can be downloaded from http://boshartifacts.cfapps.io/file_collections?type=stemcells.

### Upload Release<a name="upload_release"></a>

You can use a pre-built final release or build a release from HEAD. Final releases contain pre-compiled packages, making deployment much faster. However, these are created manually and infrequently. To be sure you're deploying the latest code, build a release yourself.

#### Upload a pre-built final BOSH release

1. Check out the tag for the desired version. This is necessary for generating a manifest that matches the code you're deploying.

  ```bash
  cd ~/workspace/cf-mysql-release
  ./update
  git checkout v8
  ```

1. Run the upload command, referencing one of the config files in the `releases` directory.

  ```bash
  bosh upload release releases/cf-mysql-8.yml
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

### Generating a Deployment Manifest<a name="create_manifest"></a>

We have provided scripts to help you generate a deployment manifest.  These scripts currently support AWS, vSphere, and [bosh-lite](https://github.com/cloudfoundry/bosh-lite) deployments.

The scripts we provide require [Spiff](https://github.com/cloudfoundry-incubator/spiff) to be installed on the local workstation.  Spiff is a tool we use to help generate a deployment manifest from "stubs", YAML files with values unique to the deployment environment (two identical deployments of Cloud Foundry will have stubs with the same keys but some unique values).  Stub files make it easier to consider only the keys/values that are important to you without having to comb through an entire deployment manifest file, which can be quite large.

To generate a deployment manifest for bosh-lite, follow the instructions [here](#using-bosh-lite).

To generate a deployment manifest for AWS or vSphere, use the [generate_deployment_manifest](generate_deployment_manifest) script.  We recommend the following workflow:

1. Run the `generate_deployment_manifest` script.
1. If you're missing manifest parameters in your stub, you'll get a list of missing manifest parameters. Check the `spec` file for each job in `jobs/#{job_name}/spec`. These spec files contain all the required parameters you will need to supply.
1. Add those paramaters and values into the stub.  See [Hints for missing parameters in your deployment manifest stub](#hints-for-missing-parameters-in-your-deployment-manifest-stub) below.
1. When all necessary stub parameters are present, the script will output the deployment manifest to stdout. Pipe this output to a file in your environment directory that indicates the environment and the release, e.g. `~/workspace/deployments/mydevenv/cf-mysql-mydevenv.yml`.

#### Example using AWS:

```bash
./generate_deployment_manifest aws ~/workspace/deployments/mydevenv/stub.yml

2013/12/16 09:57:18 error generating manifest: unresolved nodes:
    dynaml.MergeExpr{[jobs mysql properties admin_password]}
    dynaml.MergeExpr{[jobs cf-mysql-broker properties auth_username]}
    dynaml.MergeExpr{[jobs cf-mysql-broker properties auth_password]}
    dynaml.ReferenceExpr{[jobs mysql properties admin_password]}
```

These errors indicate that the deployment manifest stub is missing the following fields:

    ---
    jobs:
      mysql:
        properties:
          admin_password: <choose_admin_password>
      cf-mysql-broker:
        properties:
          auth_username:
          auth_password:


#### Hints for missing parameters in your deployment manifest stub:

Properties you will need to edit:

- `director_uuid`: Shown by running `bosh status`
- `admin_password`: The admin password for the MySQL server process. You should generate a secure password and configure it using this parameter.
- `auth_username`: The username cloud controller will use to authenticate with the service broker.
- `auth_password`: The password cloud controller will use to authenticate with the service broker.

#### For AWS:

You need to know the AZ and subnet id, and you will need to configure them in the stub:

- `availability_zone`: From the EC2 page of the AWS console, like `us-east-1a`.
- `subnet_id`:  From VPC/Subnets page of AWS console.  Availability zone must match the value set above.  

#### For vSphere:

You need the available IP address range and the IP address of the DNS server and should configure these in the stub:

- `networks`: Follow example from `templates/cf-infrastructure-aws.yml`.  Set IP addresses.  The `networks.subnets.cloud_properties` field requires only a sub-field called `name`.  This should match your vSphere network name, e.g. "VM Network".

#### Using bosh-lite

Running the [make_manifest_spiff_mysql](bosh-lite/make_manifest_spiff_mysql) script requires that you have bosh-lite installed and running on your local workstation.  Instructions for doing that can be found on the [bosh-lite README](https://github.com/cloudfoundry/bosh-lite).

For bosh-lite we provide a fully configured stub, including some default values you will need later:

- `admin_password` defaults to password.
- `auth_username` defaults to admin.
- `auth_password` defaults to password.

Run the `make_manifest_spiff_mysql` script to generate your manifest, which you can find in [cf-mysql-release/bosh-lite/](bosh-lite/).

Example:
```bash
./bosh-lite/make_manifest_spiff_mysql
# This step would have also set your deployment to ./bosh-lite/manifests/cf-mysql-manifest.yml
```

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
  "max_user_connections": 40,
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
