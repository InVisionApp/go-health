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
* [Redis](#redis)
* [SQL DB](#sql-db)
* [Mongo](#mongo)

### HTTP
The HTTP checker is a generic HTTP call executor. To make use of it, instantiate and fill out a `HTTPConfig` struct and pass it into `checkers.NewHTTP(...)`.

The only **required** attribute is `HTTPConfig.URL` (`*url.URL`). 
Refer to the source code for all available attributes on the struct.

### Redis
The Redis checker allows allows you to test that your server is either available (by ping), is able to set a value, is able to get a value or all of the above.

To make use of it, instantiate and fill out a `RedisConfig` struct and pass it to `checkers.NewRedis(...)`.

The `RedisConfig` must contain a valid `RedisAuthConfig` and at least _one_ check method (ping, set or get).

Refer to the godocs for additional info.

### SQL DB
Planned, but PR's welcome!

### Mongo
Planned, but PR's welcome!
