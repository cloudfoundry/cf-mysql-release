# Deploying with bosh 2.0

`Bosh 2.0` is a generic term referring to a set of new bosh features including:
- [cloud config](https://bosh.io/docs/cloud-config.html)
- [job links](https://bosh.io/docs/links.html)
- [new CLI](https://github.com/cloudfoundry/bosh-cli)

These features are designed to simplify managing bosh deployments, and we
strongly recommend all operators take advantage of these new features.

## Prerequisites
- The new BOSH CLI must be installed according to the instructions [here](https://bosh.io/docs/cli-v2.html).

- The new BOSH CLI requires that directors have an SSL cert and that this cert
is valid for the domain. Self-signed certs can be validated against the root
certificate authority by providing the root ca to the bosh CLI as follows:

  ```sh
  bosh \
    --ca-cert <path-to-ca-cert> \
    login
  ```

- The BOSH director must be relatively new (e.g. links was available
starting with bosh-release `v255.5`).

- Add cloud-config to the director with:

  ```sh
  bosh update-cloud-config <path-to-cloud-config>
  ```

## New deployments

New deployments will work "out of the box" with little additional configuration.
There are two mechanisms for providing credentials to the deployment:

- Credentials can be provided with `-l <path-to-vars-file>` (see below for more
information on variable files).
- variables store file should be provided with
`--vars-store <path-to-vars-store-file>` to let the CLI generate secure passwords
and write them to the provided vars store file.

By default the deployment manifest will not deploy brokers, nor try to register
routes for the proxies with a Cloud Foundry router. To enable integration with
Cloud Foundry, overrides files are provided to
[add brokers](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/bosh2.0/overrides/add-broker.yml)
and
[register proxy routes](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/bosh2.0/overrides/register-proxy-route.yml).

If you require static IPs for the proxy instance groups, these IPs should be
added to the `networks` section of the cloud-config as well as to an override file
which will use these IPs for the proxy instance groups. See below for more information on override files.

```sh
bosh \
  -e my-director \
  -d cf-mysql \
  deploy \
  ~/workspace/cf-mysql-release/manifest-generation/cf-mysql-template-v2.yml \
  -o <path-to-overrides
```

## Upgrading from previous deployment topologies

If you are upgrading an existing deployment of cf-mysql-release with a manifest
that does not take advantage of these new features, for example if the manifest
was generated via the spiff templates and stubs provided in this repository,
then you will need to follow the steps outlined below:

1. Ensure the networks defined in cloud-config match what was previously defined in the mysql manifest.
1. Create an override to set the new deployment name to your existing one. See below for instructions on creating override files.
1. By default the deployment manifest will not deploy brokers, nor try to register
routes for the proxies with a Cloud Foundry router. If you wish to preserve this
behavior you will need to include the
[add brokers](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/bosh2.0/overrides/add-broker.yml)
and
[register proxy routes](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/bosh2.0/overrides/register-proxy-route.yml) override files.
1. Create custom override files to map any non-default configuration (e.g.
the number of maximum connections).
1. Create a variables file to contain the credentials of the existing deployment.
 - Using `--vars-store` is not recommended as it will result in credentials being rotated which can cause issues.
1. Run the following command:

```sh
bosh \
  -e my-director \
  -d my-deployment \
  deploy \
  ~/workspace/cf-mysql-release/manifest-generation/cf-mysql-template-v2.yml \
  -o <path-to-deployment-name-override> \
  [-o <path-to-additional-overrides>] \
  -l <path-to-vars-file> \
  [-l <path-to-additional-vars-files>]
```

## Overrides files

Overrides files are optional files for modifying the deployment manifest.
They are intended for structural and non-secret changes, e.g. modifying the
`cf_mysql.mysql.max_connections` property. Secret values should be placed in
variables files (see below for more information on variables files).

A set of override files for common functionality (e.g. adding a broker for
Cloud Foundry integration) is provided
[here](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/bosh2.0/overrides).

The syntax for override files is detailed
[here](https://github.com/cppforlife/go-patch/blob/master/docs/examples.md),
and example overrides files can be found [here](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/examples/bosh2.0/overrides).

Override files can be provided at deploy-time as follows:

```sh
bosh \
  deploy \
  -o <path-to-overrides-file>
```

## Variables files

Variables files are a flat-format key-value yaml file which contains sensitive
information such as password, ssl keys/certs etc.

They can be provided at deploy-time as follows:

```sh
bosh \
  deploy \
  -l <path-to-vars-file>
```

An example variable file for bosh-lite can be found
[here](https://github.com/cloudfoundry/cf-mysql-release/tree/master/manifest-generation/bosh2.0/bosh-lite/default-vars.yml).
