# Cloud Foundry MySQL Service

This project contains a BOSH release of a MySQL service for Cloud Foundry. It utilizes the [v2 broker API](http://docs.cloudfoundry.com/docs/running/architecture/services/writing-service.html).

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

- The CF MySQL service requires a fully working version of [Cloud Foundry (runtime)](https://github.com/cloudfoundry/cf-release) (release [155](https://github.com/cloudfoundry/cf-release/tree/v155) or higher).
- Installing the CF MySQL service requires BOSH.
- Instructions on installing BOSH as well as Cloud Foundry (runtime) are located in the [Cloud Foundry documentation](http://docs.cloudfoundry.com/docs/running/deploying-cf/).

Steps:

1. Create a CF MySQL deployment manifest
1. (Optional) Create a BOSH release for CF MySQL
1. Upload the release to the BOSH director
1. Deploy the CF MySQL release with BOSH
1. Register the newly created service broker with the Cloud Controller
1. Make the CF MySQL plans public

The MySQL service should now be advertised when running `gcf marketplace`

### Generating a Deployment Manifest

We have provided scripts to help you generate a deployment manifest.  These scripts currently support AWS, vSphere, and [bosh-lite](https://github.com/cloudfoundry/bosh-lite) deployments.

The scripts we provide require [Spiff](https://github.com/cloudfoundry-incubator/spiff) to be installed on the local workstation.  Spiff is a tool we use to help generate a deployment manifest from "stubs", YAML files with values unique to the deployment environment (two identical deployments of Cloud Foundry will have stubs with the same keys but some unique values).  Stub files make it easier to consider only the keys/values that are important to you without having to comb through an entire deployment manifest file, which can be quite large.

To generate a deployment manifest for bosh-lite, follow the instructions [here](#using-bosh-lite).

To generate a deployment manifest for AWS or vSphere, use the [generate_deployment_manifest](generate_deployment_manifest) script.  We recommend the following workflow:

1. Run the `generate_deployment_manifest` script. You'll get some error that indicates what the missing manifest parameters are. 
1. Add those paramaters and values into the stub.  See [Hints for missing parameters in your deployment manifest stub](#hints-for-missing-parameters-in-your-deployment-manifest-stub) below.
1. Rinse and repeat
1. When all necessary stub parameters are present, the script will output the deployment manifest to stdout. Pipe this output to a file in your environment directory which indicates the environment and the release, such as `~/workspace/deployments/mydevenv/cf-mysql-mydevenv.yml`.

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

You need the available IP address range and the IP address of the DNS server, and should configure these in the stub:

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

### (Optional) Create a BOSH Release

You can build a release from HEAD, or use a pre-built final release. Final releases contain pre-compiled packages, making deployment much faster. To build the release from HEAD:

    $ ./update
    $ bosh create release
    
When prompted to name the release, called it `cf-mysql`.

### Upload Release

    $ bosh upload release

If you'd like to use a pre-built final release, reference one of the config files in the `releases` directory in your upload command. For example:

    $ bosh upload releases/cf-mysql-6.yml

The [cf-release document](http://docs.cloudfoundry.com/docs/running/deploying-cf/common/cf-release.html) provides additional details on uploading releases using BOSH.

### Deploy Using BOSH

Set your deployment using the deployment manifest you generated above.

    $ bosh deployment ~/workspace/deployments/mydevenv/cf-mysql-mydevenv.yml
    $ bosh deploy
    
If you followed the instructions for bosh-lite above your manifest is in the `cf-mysql-release/bosh-lite/manifests` directory. The make\_manifest\_spiff\_mysql script should have already set the deployment to the manifest, so you just have to run:

    $ bosh deploy

[Deploying Cloud Foundry with BOSH](http://docs.cloudfoundry.com/docs/running/deploying-cf/vsphere/deploy_cf_vsphere.html) provides additional details on deploying with BOSH.

### Register the CF MySQL Service Broker

1. Target Cloud Foundry and login as an admin user
    
    If you're using bosh-lite, you may have to run this script first:
    
    ```
    $ ~/workspace/bosh-lite/scripts/add-route
    ```
    
2. Run the following command to register the MySQL broker

    ```
    $ gcf create-service-broker p-mysql BROKER_USERNAME BROKER_PASSWORD URL
    ```
    
    - BROKER_USERNAME and BROKER_PASSWORD are the values you gave for `auth_username` and `auth_password` in the deployment manifest. 
    - URL specifies where the Cloud Controller will access the MySQL broker. If DNS is not configured for the MySQL broker, specify a URL using the IP address such as `http://10.244.1.130`. You can discover the broker IP address with the BOSH command, `bosh vms`.
    
For more information, see the documentation on [Managing Service Brokers](http://docs.cloudfoundry.com/docs/running/architecture/services/managing-service-brokers.html).

### Make MySQL Service Plan Public

By default new plans are private, which means they are not visible to end users. This enables an admin to test services before making them available to end users.

To make a plan public, use the old ruby CF CLI (the curl feature will be implemented soon on gcf).


1. Login as an admin user with the ruby cf cli.

    ```
    $ cf login
    ```

2. Get the service plan guid.

    ```
    $ cf services -m -t
    ```

   This returns a JSON response which includes something like the following for each service.

    ```
    "service_plans": [
      {
        "metadata": {
          "guid": "a01d462f-a4a4-4945-a008-3ff13c06f719",
          "url": "/v2/service_plans/a01d462f-a4a4-4945-a008-3ff13c06f719",
          "created_at": "2013-11-15T23:42:52+00:00",
          "updated_at": "2013-11-22T18:57:46+00:00"
        },
        "entity": {
          "name": "large",
          "free": false,
          "description": "Large Dummy",
          "service_guid": "e629bb0a-fee7-4a6c-a4f1-9eeec7096c29",
          "extra": "{\"cost\":20.0,\"bullets\":[]}",
          "unique_id": "addonOffering_3365",
          "public": false,
          "service_url": "/v2/services/e629bb0a-fee7-4a6c-a4f1-9eeec7096c29",
          "service_instances_url": "/v2/service_plans/a01d462f-a4a4-4945-a008-3ff13c06f719/service_instances"
        }
      }
    ]      
    ```

3. Set the plan as public, using its GUID.

    ```
    $ cf curl PUT /v2/service_plans/a01d462f-a4a4-4945-a008-3ff13c06f719 -b '{"public":'true'}'
    ```

4. Verify that the the plan was set to public. Re-run this command and check the 'public' field:

    ```
    $ cf services -m -t
    ```
    
    You should now be able to view the p-mysql service in the service marketplace and create service instances.

    ```
    $ gcf m
    Getting services from marketplace in org scoen / space broker as admin...
    OK
    service   plans   description
    p-mysql   100mb   MySQL service for application development and testing
    ```



