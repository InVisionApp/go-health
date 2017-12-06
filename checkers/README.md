checkers
========
The `health` library comes with a number of built-in checkers for well known
types of dependencies.

If a pre-built checker is not available, you can create your own checkers by
implementing the `ICheckable` interface (which consists of a single method - 
`Status() (interface{}, error)`).

If you do create a custom-checker - consider opening a PR and adding it to the
list of built-in checkers.

## Built-in checkers

* [HTTP](#http)
* [DB Ping](#db-ping)
* [DB SQL](#db-sql)
* [Mongo](#mongo)
* [Redis](#redis)

### HTTP
The HTTP checker is a generic HTTP call executor. To make use of it, instantiate and fill out a `HTTPConfig` struct and pass it into `checkers.NewHTTP(...)`.

The only **required** attribute is `HTTPConfig.URL` (`*url.URL`). 
Refer to the source code for all available attributes on the struct.

### DB Ping
WIP

### DB SQL
WIP

### Mongo
WIP

### Redis
WIP