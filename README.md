# gobdns

A simple, dynamic dns server written in go. It can direct all requests for an
address, or addresses matching `*.<address>` or `*-<address>` to a particular
ip.

For example, if you set `brian.turtles.com` to point to `127.0.0.5`, then all
these domains will get directed at that address:

* `brian.turtles.com`
* `foo.brian.turtles.com`
* `bar-brian.turtles.com`
* `foo-bar.brian.turtles.com`

and so on. This is useful for setting up development environments for multiple
employees, each one maybe having multiple virtual hosts.

## Features

* TCP and UDP support
* Easy to setup
* Prefix-wildcard matching of requests
* Simple REST api for retrieving and modifying entries
* Disk persistance
* Basic master/slave replication
* Can forward requests for unknown domains to a different nameserver

## Building

First, clone the repo. Then, if you don't have the dependencies yet, simply:

    go get ./...

After that:

    go build

To build the gobdns binary

## Usage

### Running

Run `./gobdns --example` to output an example configuration file. This can be
piped to a file, modified and used with the `--config` flag.

### Adding entries

Once running, you can use the REST api to add and remove entries (TODO: hoping
for an actual web console soon).

Assuming the REST interface is on `localhost:8080`:

See all existing entries:

    curl -i localhost:8080/api/domains/all

Set `brian.turtles.com` to point to `127.0.0.5`:

    curl -i -XPOST -d'127.0.0.5' localhost:8080/api/domains/brian.turtles.com

Delete the `brian.turtles.com` entry:

    curl -i -XDELETE localhost:8080/api/domains/brian.turtles.com

## Persistance

If `backup-file` is set (it is by default) then every second a snapshot of the
current data set will be written to disk. On startup (again, if `backup-file` is
set) this file will be read in if it exists and used as the initial set of
mappings.

## Replication

`master-addr` can be set to point to the REST interface of another running
gobdns instance, and every 5 seconds will pull the full list of entries from
that instance and overwrite the current list.

## TODO

* consolidate some duplicated code in persist and repl
* make an Overwrite method in ips, and use that in repl so that deleted entries
  go away