## Acceptance Tests

The acceptance tests are for developers to validate changes to the MySQL Release.

To run the MySQL Release Acceptance tests, you will need:
- a running CF instance
- credentials for a CF Admin user
- a deployed MySQL Release with the broker registered and the plan made public

### BOSH errand

BOSH errands were introduced in version 2366 of the BOSH CLI, BOSH Director, and stemcells.

The acceptance tests requires the same deployment manifest properties as the [smoke tests](/README.md#running-smoke-tests-via-bosh-errand).

By default, the acceptance-tests errand only runs the smoke tests. To enable the full acceptance test suite, set the following property in the BOSH manifest:

```
jobs:
- name: acceptance-tests
  properties:
    smoke_tests_only: false
```

To run the acceptance tests via bosh errand:

```
$ bosh run errand acceptance-tests
```

### Manually

The acceptance tests can also be run manually. One advantage to doing this is that output will be streamed in real-time, as opposed to the output from a bosh errand output which is printed all at once when it finishes running.

Instructions:

1. Install **Go** by following the directions found [here](http://golang.org/doc/install)
1. Export the environment variable `$GOPATH` set to the `cf-mysql-release` directory to manage Golang dependencies. For more information, see [here](https://github.com/cloudfoundry/cf-mysql-release/tree/release-candidate#development).
1. Change to the acceptance-tests directory:

    ```
    $ cd cf-mysql-release/src/github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/
    ```

1. Install [Ginkgo](http://onsi.github.io/ginkgo/):

    ```
    $ go get github.com/onsi/ginkgo/ginkgo
    ```

1. Configure the tests by creating `integration_config.json` and setting the environment variable `$CONFIG` to point to it. The following commands provide a shortcut to configuring `integration_config.json` with values for a [bosh-lite](https://github.com/cloudfoundry/bosh-lite)
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

1. Run the smoke tests:

  ```
  ./bin/test-smoke
  ```

1. Run the acceptance tests:

  ```
  ./bin/test-acceptance
  ```

### Dashboard Tests

The dashboard tests mostly test authentication through single sign on (SSO).

They are most easily run from within the [cf-mysql-ci docker container](https://registry.hub.docker.com/u/cloudfoundry/cf-mysql-ci/), using the provided scripts.

It is expected that the `integration_config.json` is already present in the `cf-mysql-release/src/github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/` directory.

1. From the `cf-mysql-release` directory, run the dashboard tests inside the docker container:

  ```
  cd ~/workspace/cf-mysql-release
  ./scripts/ci/run_in_docker ./scripts/test-dashboard
  ```
