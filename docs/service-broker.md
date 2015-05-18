# Service Broker

The service broker handles the service instance lifecycle. This includes creating, updating and deleting the service instance, as well as creating, updating and deleting service instance bindings.

## Service instance capacity allocation

When the service broker receives a request to create a service instance, it calculates whether there is enough space on the persistent disk to honor the request and create the service instance.

The amount of space required is the sum of the requested plan size plus a small amount of overhead.

If there is sufficient available capacity for the service instance the broker allocates space for the entire service.

If there is insufficient available capacity the broker rejects the request with HTTP response code 507 (`Insufficient Storage`) and no space is allocated.

As the service broker allocates space for the entire service instance upon request, it does not allow any form of oversubscription.
