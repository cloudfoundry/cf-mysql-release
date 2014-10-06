# MySQL Test App

This app is used by the acceptance tests and is an example of how you can use the service from an app.

The app allows you to write and read data in the form of key:value pairs to a Riak CS bucket using RESTful endpoints.

### Getting Started

Install the app by pushing it to your Cloud Foundry and binding with the Riak CS service

Example:

    $ cf push mysqltest --no-start
    $ cf create-service p-mysql 100mb-dev mydb
    $ cf bind-service mysqltest mydb
    $ cf restart mysqltest

### Endpoints

#### PUT /:key

Stores a key:value pair in the MySQL database. Example:

    $ curl -X POST mysqltest.my-cloud-foundry.com/service/mysql/mydb/foo -d 'bar'
    success


#### GET /:key

Returns the value stored in the database for a specified key. Example:

    $ curl -X GET mysqltest.my-cloud-foundry.com/service/mysql/mydb/foo
    bar

