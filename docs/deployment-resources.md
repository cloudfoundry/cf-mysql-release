# Deployment resources

By default, the service is configured with a small footprint, suitable for development but possibly insufficient for production workloads. This document describes symptoms of an under-resourced deployment, and steps that may be taken to reconfigure the service to increase available resources.

## Symptoms

* The database VMs may have a high percentage of CPU and/or memory usage
* The persistent disk attached to the database VMs may fill up

## Configuration

The resource configuration is defined in the manifest under `resource_pools`. Currently resource configuration is currently specific to each cloud provider (AWS, vsphere etc) via the `cloud_properties` property.

### AWS

AWS provides instance types which map to specific values of CPU,RAM and ephemeral disk; they are not independently configurable. The instance type is set via the property `cloud_properties.instance_type`

If the manifest is being generated using spiff and merging the provided templates and stubs, a stub can be provided with the following structure:

```yml
---
resource_pools:
- name: services-small
  cloud_properties:
    instance_type: <some-value>
```

### VSphere

VSphere allows configuration of the CPU, RAM and ephemeral disk independently; these can be set via `cloud_properties.cpu`,`cloud_properties.ram` and `cloud_properties.disk` respectively.

If the manifest is generated using spiff, merging the provided templates and stubs, a stub can be provided with the following structure:

```yml
---
resource_pools:
- name: services-small
  cloud_properties:
    cpu: <some-value>
    ram: <some-value>
    disk: <some-value>
```
