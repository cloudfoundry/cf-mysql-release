# cf-mysql-quota-enforcer
Quota enforcer for cf-mysql-release

## Usage

### Configuration

The quota enforcer executable requires either a config file or a config string, encoded in json.
[service-config](https://github.com/pivotal-cf-experimental/service-config) is used to load the config.

Examples:
- `$ cf-mysql-quota-enforcer -configPath=/path/to/config.json`
- `$ cf-mysql-quota-enforcer -config='{"Host": "127.0.0.1", "Port": 3306, "User": "root", "Password": "password", "DBName": "development", "PauseInSeconds": 1}'`


An example configuration file is provided in `config-example.yaml`.
Copy this to `config.yaml` and edit as necessary; `config.yaml` is ignored by git.

##Testing

Unit tests can be run by executing

```sh
./bin/test-unit
```

Integration tests can be run by executing

```sh
./bin/test-integration
```

Configuration for the integration tests is managed by environment variables; see
`./bin/test-integration` for further details.
