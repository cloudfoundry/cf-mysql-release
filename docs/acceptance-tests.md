## Acceptance Tests

The acceptance tests are for developers to validate changes to the MySQL Release.

To run the MySQL Release Acceptance tests, you will need:
- a running CF instance
- credentials for a CF Admin user
- a deployed MySQL Release with the broker registered and the plan made public
- an environment variable `$CONFIG` which points to a `.json` file that contains the application domain

### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

The acceptance tests requires the same deployment manifest properties as the [smoke tests](/README.md#running-smoke-tests-via-bosh-errand).

By default, the acceptance-tests errand only runs the smoke tests. To enable the full acceptance test suite, set the following property:

- `smoke_tests_only: false`

To run the acceptance tests via bosh errand:

```
$ bosh run errand acceptance-tests
```

### Manually

The acceptance tests can also be run manually.

Reasons to run tests manually:

1. Output will be streamed (bosh errand output is printed all at once when they finish running)

Instructions:

1. Install **Go** by following the directions found [here](http://golang.org/doc/install)
2. `cd` into `cf-mysql-release/src/acceptance-tests/`
3. Update `cf-mysql-release/src/github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/integration_config.json`

    The following commands provide a shortcut to configuring `integration_config.json` with values for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite)
deployment. Copy and paste this into your terminal, then open the resulting `integration_config.json` in an editor to replace values as appropriate for your environment.

  ```bash
  cat > integration_config.json <<EOF
  {
    "api": "http://api.10.244.0.34.xip.io",
    "apps_domain": "10.244.0.34.xip.io",
    "admin_user": "admin",
    "admin_password": "admin",
    "broker_host": "p-mysql.10.244.0.34.xip.io",
    "service_name": "p-mysql",
    "plans" : [
      {
        "plan_name": "100mb",
        "max_user_connections": 20,
        "max_storage_mb": 10
      },
      {
        "plan_name": "1gb",
        "max_user_connections": 40,
        "max_storage_mb": 20
      }
    ],
    "skip_ssl_validation": true,
    "timeout_scale": 1.0,
    "smoke_tests_only": false,
    "include_dashboard_tests": false,
    "include_failover_tests": false,
    "proxy": {
      "external_host":"p-mysql.10.244.0.34.xip.io",
      "api_username":"username",
      "api_password":"password"
    }
  }
  EOF
  export CONFIG=$PWD/integration_config.json
  ```

  When `skip_ssl_validation: true`, commands run by the tests will accept self-signed certificates from Cloud Foundry. This option requires v6.0.2 or newer of the CLI.

  All timeouts in the test suite can be scaled proportionally by changing the `timeout_scale` factor.

  To run only the smoke tests, use `"smoke_tests_only": true` in the above config.

4. Run  the tests

  ```
  $ ./bin/test
  ```
