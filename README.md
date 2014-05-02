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

1. [Create a CF MySQL deployment manifest](#create_manifest)
1. [(Optional) Create a BOSH release for CF MySQL](#create_release)
1. [Upload the release to the BOSH director](#upload_release)
1. [Deploy the CF MySQL release with BOSH](#deploy_release)
1. [Register the newly created service broker with the Cloud Controller](#register_broker)
1. [Make the CF MySQL plans public](#publicize_plans)

The MySQL service should now be advertised when running `gcf marketplace`

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

    $ ./generate_deployment_manifest aws ~/workspace/deployments/mydevenv/stub.yml

    2013/12/16 09:57:18 error generating manifest: unresolved nodes:
	    dynaml.MergeExpr{[jobs mysql properties admin_password]}
	    dynaml.MergeExpr{[jobs cf-mysql-broker properties auth_username]}
	    dynaml.MergeExpr{[jobs cf-mysql-broker properties auth_password]}
	    dynaml.ReferenceExpr{[jobs mysql properties admin_password]}

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
```
$ ./bosh-lite/make_manifest_spiff_mysql
# This step would have also set your deployment to ./bosh-lite/manifests/cf-mysql-manifest.yml
```

### (Optional) Create a BOSH Release<a name="create_release"></a>

You can build a release from HEAD, or use a pre-built final release. Final releases contain pre-compiled packages, making deployment much faster. To build the release from HEAD:

    $ ./update
    $ bosh create release
    
When prompted to name the release, called it `cf-mysql`.

### Upload Release<a name="upload_release"></a>

    $ bosh upload release

If you'd like to use a pre-built final release, reference one of the config files in the `releases` directory in your upload command. For example:

    $ bosh upload release releases/cf-mysql-6.yml

### Deploy Using BOSH<a name="deploy_release"></a>

Set your deployment using the deployment manifest you generated above.

    $ bosh deployment ~/workspace/deployments/mydevenv/cf-mysql-mydevenv.yml
    $ bosh deploy
    
If you followed the instructions for bosh-lite above, your manifest is in the `cf-mysql-release/bosh-lite/manifests` directory. The make\_manifest\_spiff\_mysql script should have already set the deployment to the manifest, so you just have to run:

    $ bosh deploy

### Register the CF MySQL Service Broker<a name="register_broker"></a>

### Using BOSH errands

If you're using a new enough BOSH director, stemcell, and CLI to support errands, run the following errand:

        bosh run errand broker-registrar
        
Note: the broker-registrar errand will fail if the broker has already been registered, and the broker name does not match the manifest property `jobs.broker-registrar.properties.broker.name`. Use the `cf rename-service-broker` CLI command to change the broker name to match the manifest property then this errand will succeed. 

### Manually

1. First register the broker using the `cf` CLI.  You must be logged in as an admin.

    ```
    $ cf create-service-broker p-mysql BROKER_USERNAME BROKER_PASSWORD URL
    ```
    
    - `BROKER_USERNAME` and `BROKER_PASSWORD` are the credentials Cloud Foundry will use to authenticate when making API calls to the service broker. Use the values you gave for the manifest properties `jobs.cf-mysql-broker.properties.auth_username` and `jobs.cf-mysql-broker.properties.auth_password`. 
    - `URL` specifies where the Cloud Controller will access the MySQL broker. Use the value of the manifest property `jobs.cf-mysql-broker.properties.external_host`.
    For more information, see [Managing Service Brokers](http://docs.cloudfoundry.org/services/managing-service-brokers.html).

2. Then [make the service plan public](http://docs.cloudfoundry.org/services/services/managing-service-brokers.html#make-plans-public).

### New Features in this Release (v7)
#### Errands
##### Acceptance Tests
This release includes an errand to run acceptance tests. They can be run using this command:

`$ bosh run errand acceptance-tests`

##### De-registration of the CF MySQL Service Broker

This bosh release also has an errand to de-register the broker and purge all services/service instances along with it. To do this, simply run:

`$ bosh run errand broker-deregistrar`

This is equivalent to running the following CLI commands:

    $ cf purge-service-offering p-mysql
    $ cf delete-service-broker p-mysql

#### Dashboard <a name="dashboard"></a>

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

