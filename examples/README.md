### Run the example

To spin up the example servers please run 
```shell
make run-nodes
```

Any testing you wish to do can be made with a cURL (or anything you want)

The item type is thought of to be `string` in this example.

to run an automated example please run `make run-nodes` and in another shell

```shell
make example
```

### Walkthrough

We have two servers spin up (and register their peer manually) and then we are free to add and remove elements as we see fit and then get the synced list of elements with high consistency from any of the servers.

The idea here is to add a couple of items to one server, than you can reove one of them from the other server. Upon asking for the items you will always receive the correct list, from either server.

This is achieved with use of the CRDT data type as an abstraction to a data store.

### Bibliography

I want to give credit for the Peer Sync logic to [el10savio](https://github.com/el10savio/twoPSet-crdt) from whom I took a lot of inspiration to write this example.