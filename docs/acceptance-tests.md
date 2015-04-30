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

```yml
jobs:
- name: acceptance-tests
  properties:
    smoke_tests_only: false
```

By default, the acceptance tests will create a random organization and delete it after each test run.
The tests assume that all plans are public, and will fail if any plans are private.
To run the tests against private plans, create an organization to be used by the acceptance tests which has access to all plans. Ensure the organization has a correctly configured quota; see [quota documentation](http://docs.cloudfoundry.org/running/managing-cf/quota-plans.html) for more details.

```
cf create-org MY_TEST_ORG
cf enable-service-access p-mysql -o MY_TEST_ORG
```

Then add the following property to the deployment manifest and (re)deploy:
```yml
jobs:
- name: acceptance-tests
  properties:
    org_name: MY_TEST_ORG
```

To run the acceptance tests via bosh errand:

```bash
$ bosh run errand acceptance-tests
```

### Manually

The acceptance tests can also be run manually. One advantage to doing this is that output will be streamed in real-time, as opposed to the output from a bosh errand output which is printed all at once when it finishes running.

Instructions:

1. Install **Go** by following the directions found [here](http://golang.org/doc/install)
1. Export the environment variable `$GOPATH` set to the `cf-mysql-release` directory to manage Golang dependencies. For more information, see [here](https://github.com/cloudfoundry/cf-mysql-release/tree/release-candidate#development).
1. Change to the acceptance-tests directory:

    ```bash
$ cd cf-mysql-release/src/github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/
    ```

1. Install [Ginkgo](http://onsi.github.io/ginkgo/):

    ```bash
$ go get github.com/onsi/ginkgo/ginkgo
    ```

1. Configure the tests.

  Create a config file and set the environment variable `$CONFIG` to point to it. For bosh-lite, this can easily be achieved by executing the following command and following the instructions on screen:

    ```bash
$ ~/workspace/cf-mysql-release/bosh-lite/create_integration_test_config
    ```

 Open the resulting file in an editor to replace values as appropriate for your environment.

  When `skip_ssl_validation: true`, commands run by the tests will accept self-signed certificates from Cloud Foundry. This option requires v6.0.2 or newer of the CLI.

  All timeouts in the test suite can be scaled proportionally by changing the `timeout_scale` factor.

1. Run the smoke tests:

  ```bash
$ ./bin/test-smoke
  ```

1. Run the acceptance tests:

  ```bash
$ ./bin/test-acceptance
  ```

### Dashboard Tests

The dashboard tests validate the ability of a user to navigate to the broker dashboard via single sign on (SSO) authentication.

They are most easily run from within the [cf-mysql-ci docker container](https://registry.hub.docker.com/u/cloudfoundry/cf-mysql-ci/), using the provided scripts.

The config file must already be configured as detailed above.

1. From the `cf-mysql-release` directory, run the dashboard tests inside the docker container:

  ```bash
$ cd ~/workspace/cf-mysql-release
$ ./scripts/ci/run_in_docker ./scripts/test-dashboard
  ```
