# New deployment procedure for cf-mysql-release

 
## Getting Started

* [bosh go CLI](https://github.com/cloudfoundry/bosh-cli): `brew install chendrix/tap/gobosh`
* [cf-mysql-release](https://github.com/cloudfoundry/cf-mysql-release) repository
* Prepare a [cloud-config](http://bosh.io/docs/cloud-config.html) and update your environment to use it


## Upload Release

`gobosh -e <environment> upload-release https://bosh.io/d/github.com/cloudfoundry/cf-mysql-release`


## Prepare Variables File

 We provide a default set of variables intended for a local bosh-lite environment in a var-file called 
 [`cf-mysql-release/manifest-generation/bosh2.0/bosh-lite/default-vars.yml`](https://github.com/cloudfoundry/cf-mysql-release/blob/develop/manifest-generation/bosh2.0/bosh-lite/default-vars.yml)

Use this as an example for your environment-specific var-file, e.g. `cf-mysql-vars.yml`


## Prepare Operations File

Any changes to our provided [v2 schema manifest template](https://github.com/cloudfoundry/cf-mysql-release/blob/develop/manifest-generation/cf-mysql-template-v2.yml) 
should be added to an environment-specific ops-file, e.g. `cf-mysql-overrides.yml`

For instance, our template assumes your cloud-config has a persistent-disk type named "large". If your cloud-config specifies a type "small", you would add the following to `cf-mysql-overrides.yml`:

```yml
---
- type: replace
  path: /instance_groups/name=mysql/persistent_disk_type
  value: small
```


## Deploy

```bash
gobosh \
  -e <environment> \
  deploy \
  "${CF_MYSQL_RELEASE_DIR}/manifest-generation/cf-mysql-template-v2.yml" \
  -l "cf-mysql-vars.yml" \
  -o "cf-mysql-overrides.yml"
```


