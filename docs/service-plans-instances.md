# Service plans and instances

Service plans are defined in the manifest via properties on the `cf-mysql-broker` job, for example:

```yml
---
jobs:
- cf-mysql-broker_z1:
  properties:
    plans:
    - id: unique-id
      name: 10MB
      max_storage_mb: 10
    - id: another-unique-id
      name: 100MB
      max_storage_mb: 100
    - id: some-other-unique-id
      name: 1GB
      max_storage_mb: 1000
```

Additional fields may be required or optional; refer to `jobs/cf-mysql-broker/spec` for more details.

## Updating service instances

Service instances may be updated via the CLI as follows:

```sh
cf update-service SERVICE_INSTANCE -p NEW_PLAN
```

Updating a service instance between plans behaves as follows:

* Updating a service instance to a plan with a larger `max_storage_mb` is always supported.
* Updating a service instance to a plan with a smaller `max_storage_mb` is supported only if the current usage is less than the new value.
 * If the current usage is greater than the new value the update command will fail.
