# Service Broker

The service broker handles the service instance lifecycle. This includes creating, updating and deleting the service instance, as well as creating, updating and deleting service instance bindings.

## Service instance capacity allocation

When the service broker receives a request to create a service instance, it calculates whether there is enough space on the persistent disk to honor the request and create the service instance.

The amount of space required is the sum of the requested plan size plus a small amount of overhead.

If there is sufficient available capacity for the service instance the broker allocates space for the entire service.

If there is insufficient available capacity the broker rejects the request with HTTP response code 507 (`Insufficient Storage`) and no space is allocated.

As the service broker allocates space for the entire service instance upon request, it does not allow any form of oversubscription.

## Scaling the number of Service Brokers in a deployment

We recommend that a minimum of 2 service brokers are configured for a given deployment. The Cloud Foundry router distributes load using a [round robin](http://en.wikipedia.org/wiki/Round-robin_scheduling) strategy.

Having redundant service brokers avoids a single point of failure when creating or binding service instances for an application. While a deployment is operable with a single service broker it is not recommended for stated reasons.

If zero service brokers are deployed, it follows that service instances cannot be provisioned nor bound to any application.

## Quota Enforcement

The service broker is deployed with a quota-enforcer process, which ensures that service instances do not exceed their allocated storage quota. When the quota is exceeded, the database users associated with the service instance will only be able to `DELETE` until the disk usage falls under the quota.

By default, the `roadmin` and `quota-enforcer` database users are not subject to quota enforcement. When deploying the service broker you can optionally specify additional users to ignore with the property `cf_mysql.broker.quota_enforcer.ignored_users`. For an example of this, see the [example property-overrides.yml file](../manifest-generation/examples/property-overrides.yml).

### Configuring the pause time

By default the quota enforcer pauses for 1 second between checks. This pause time is configurable through the `broker.quota_enforcer.pause` property, where the value is specified as the number of seconds to pause for.

For most deployments we recommend that the pause time be left at its default value to ensure that quotas are enforced in a timely manner.

One reason why an operator may choose to increase the pause time is when debugging Galera clustering. If `wsrep_debug` is enabled then the queries issued by the quota enforcer may obscure other activity - reducing the frequency of quota enforcer checks may be helpful in this case.
