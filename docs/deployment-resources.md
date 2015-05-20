# Deployment resources

By default, the service is configured with a small footprint, suitable for development but possibly insufficient for production workloads. This document describes symptoms of an under-resourced deployment, and steps that may be taken to reconfigure the service to increase available resources.

## Symptoms and Possible Causes

* The service is sluggish or unresponsive
  * The VMs may have a high utilization of CPU and/or memory
* The service returns "No space left on device" errors
  * The persistent or emphemeral disks attached to the database VMs may fill up

If you experience the above symptoms, check your VMs for high resource utilization. You can increase resource capacity as defined below.

## Configuration

The resource configuration is defined in the manifest under `resource_pools`. Currently resource configuration is specific to each cloud provider (AWS, vsphere etc) via the `cloud_properties` property.

### AWS

AWS provides instance types which map to specific values of CPU, RAM and ephemeral disk; they are not independently configurable. The instance type is set via the property `cloud_properties.instance_type`. If the available space on the attached ephemeral disk is frequently being consumed, consider configuring a larger instance type or attaching a larger empheral disk. A larger ephemeral disk can be attached to a given EC2 instance as shown [here](https://bosh.io/docs/aws-cpi.html#resource-pools).  Persistent disk can also be resized [as shown](https://bosh.io/docs/aws-cpi.html#disk-pools).

If the manifest is being generated using spiff and merging the provided templates and stubs, a stub can be provided with the following structure:

```yml
---
resource_pools:
- name: <job-name e.g. mysql_z1>
  cloud_properties:
    instance_type: <some-value>
```

### VSphere

VSphere allows configuration of the CPU, RAM and ephemeral disk independently; these can be set via `cloud_properties.cpu`,`cloud_properties.ram` and `cloud_properties.disk` respectively.

If the manifest is generated using spiff, merging the provided templates and stubs, a stub can be provided with the following structure:

```yml
---
resource_pools:
- name: <job-name e.g. mysql_z1>
  cloud_properties:
    cpu: <some-value>
    ram: <some-value>
    disk: <some-value>
```

### MySQL Properties

To allow finer grained control over ephemeral disk utilization, the MySQL service provides the following properties:

```yml
---
jobs:
  - name: mysql_z1
    properties:
      max_heap_table_size: 16777216
      tmp_table_size: 33554432
```

Further details on MySQL/MariaDB/s memory and temp. space usage may be found on their [website](https://mariadb.com/kb/en/mariadb/server-system-variables/#max_heap_table_size).
