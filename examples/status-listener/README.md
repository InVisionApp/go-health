## Status Listener Example

The `IStatusListener` interface allows you to hook into health check failures and recoveries as they occur.  This example runs a dependency service, `dependency.go`, and a dependent service, `service.go`.

The web server in `dependency.go` will return a `200` status code for 10 contiguous requests, then return `500` for 5 contiguous requests.  The request cycle then resets.

The web server in `service.go` uses go-health to check the dependency server.  It also uses an implementation of `IStatusListener`, which includes functions

* `HealthCheckFailed(entry *health.State)`
* `HealthCheckRecovered(entry *health.State, recordedFailures int64, failureDurationSeconds float64)`

The function `HealthCheckFailed` is triggered when the health check fails for the first time.  A count of contiguous failures will be kept until the dependency recovers.  Once the dependency does recover, the function `HealthCheckRecovered` is triggered, which reports how many healthchecks failed, as well as how long (in seconds) the dependency was in an unhealthy state.



### To run example

Within the project folder `/examples/status_listener/dependency`, run `go run dependency.go` on one terminal window, then from within `/examples/status_listener/service`, run `go run service.go` in another.  You will observe the requests to the dependency, as well as the triggering and recovery of failed health checks.

