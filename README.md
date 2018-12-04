[![LICENSE](https://img.shields.io/badge/license-MIT-orange.svg)](LICENSE)
[![Build Status](https://travis-ci.org/InVisionApp/go-health.svg?branch=master)](https://travis-ci.org/InVisionApp/go-health)
[![codecov](https://codecov.io/gh/InVisionApp/go-health/branch/master/graph/badge.svg?token=hhqA1l88kx)](https://codecov.io/gh/InVisionApp/go-health)
[![Go Report Card](https://goreportcard.com/badge/github.com/InVisionApp/go-health)](https://goreportcard.com/report/github.com/InVisionApp/go-health)
[![Godocs](https://img.shields.io/badge/golang-documentation-blue.svg)](https://godoc.org/github.com/InVisionApp/go-health)

<img align="right" src="images/go-health.svg" width="200">

# go-health
A library that enables *async* dependency health checking for services running on an orchestrated container platform such as kubernetes or mesos.

## Why is this important?
Container orchestration platforms require that the underlying service(s) expose a "health check" which is used by the platform to determine whether the container is in a good or bad state.

While this can be achieved by simply exposing a `/status` endpoint that performs synchronous checks against its dependencies (followed by returning a `200` or `non-200` status code), it is not optimal for a number of reasons:

* **It does not scale**
    + The more dependencies you add, the longer your health check will take to complete (and potentially cause your service to be killed off by the orchestration platform).
    + Depending on the complexity of a given dependency, your check may be fairly involved where it is _okay_ for it to take `30s+` to complete.
* **It adds unnecessary load on yours deps or at worst, becomes a DoS target**
    + **Non-malicious scenario**
        + Thundering herd problem -- in the event of a deployment (or restart, etc.), all of your service containers are likely to have their `/status` endpoints checked by the orchestration platform as soon as they come up. Depending on the complexity of the checks, running that many simultaneous checks against your dependencies could cause at worst the dependencies to experience problems and at minimum add unnecessary load.
        + Security scanners -- if your organization runs periodic security scans, they may hit your `/status` endpoint and trigger unnecessary dep checks.
    + **Malicious scenario**
        + Loading up any basic HTTP benchmarking tool and pointing it at your `/status` endpoint could choke your dependencies (and potentially your service).

With that said, not everyone _needs_ asynchronous checks. If your service has one dependency (and that is unlikely to change), it is trivial to write a basic, synchronous check and it will probably suffice.

However, if you anticipate that your service will have several dependencies, with varying degrees of complexity for determining their health state - you should probably think about introducing asynchronous health checks.

## How does this library help?
Writing an async health checking framework for your service is not a trivial task, especially if Go is not your primary language.

This library:

* Allows you to define how to check your dependencies.
* Allows you to define warning and fatal thresholds.
* Will run your dependency checks on a given interval, in the background. **[1]**
* Exposes a way for you to gather the check results in a *fast* and *thread-safe* manner to help determine the final status of your `/status` endpoint. **[2]**
* Comes bundled w/ [pre-built checkers](/checkers) for well-known dependencies such as `Redis`, `Mongo`, `HTTP` and more.
* Makes it simple to implement and provide your own checkers (by adhering to the checker interface).
* Allows you to trigger listener functions when your health checks fail or recover using the [`IStatusListener` interface](#oncomplete-hook-vs-istatuslistener).
* Allows you to run custom logic when a specific health check completes by using the [`OnComplete` hook](#oncomplete-hook-vs-istatuslistener).

**[1]** Make sure to run your checks on a "sane" interval - ie. if you are checking your
Redis dependency once every five minutes, your service is essentially running _blind_
for about 4.59/5 minutes. Unless you have a really good reason, check your dependencies
every X _seconds_, rather than X _minutes_.

**[2]** `go-health` continuously writes dependency health state data and allows
you to query that data via `.State()`. Alternatively, you can use one of the
pre-built HTTP handlers for your `/healthcheck` endpoint (and thus not have to
manually inspect the state data).

## Example

For _full_ examples, look through the [examples dir](examples/)

1. Create an instance of `health` and configure a checker (or two)

```golang
import (
	health "github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/checkers"
	"github.com/InVisionApp/go-health/handlers"
)

// Create a new health instance
h := health.New()

// Create a checker
myURL, _ := url.Parse("https://google.com")
myCheck, _ := checkers.NewHTTP(&checkers.HTTPConfig{
    URL: myURL,
})
```

2. Register your check with your `health` instance

```golang
h.AddChecks([]*health.Config{
    {
        Name:     "my-check",
        Checker:  myCheck,
        Interval: time.Duration(2) * time.Second,
        Fatal:    true,
    },
})
```

3. Start the health check

```golang
h.Start()
```

From here on, you can either configure an endpoint such as `/healthcheck` to use a built-in handler such as `handlers.NewJSONHandlerFunc()` or get the current health state of all your deps by traversing the data returned by `h.State()`.

## Sample /healthcheck output
Assuming you have configured `go-health` with two `HTTP` checkers, your `/healthcheck`
output would look something like this:

```json
{
    "details": {
        "bad-check": {
            "name": "bad-check",
            "status": "failed",
            "error": "Ran into error while performing 'GET' request: Get google.com: unsupported protocol scheme \"\"",
            "check_time": "2017-12-30T16:20:13.732240871-08:00"
        },
        "good-check": {
            "name": "good-check",
            "status": "ok",
            "check_time": "2017-12-30T16:20:13.80109931-08:00"
        }
    },
    "status": "ok"
}
```

## Additional Documentation
* [Examples](/examples)
  * [Status Listeners](/examples/status-listener)
  * [OnComplete Hook](/examples/on-complete-hook)
* [Checkers](/checkers)

## OnComplete Hook VS IStatusListener
At first glance it may seem that these two features provide the same functionality. However, they are meant for two different use cases:

The `IStatusListener` is useful when you want to run a custom function in the event that the overall status of your health checks change. I.E. if `go-health` is currently checking the health for two different dependencies A and B, you may want to trip a circuit breaker for A and/or B. You could also put your service in a state where it will notify callers that it is not currently operating correctly. The opposite can be done when your service recovers.

The `OnComplete` hook is called whenever a health check for an individual dependency is complete. This means that the function you register with the hook gets called every single time `go-health` completes the check. It's completely possible to register different functions with each configured health check or not to hook into the completion of certain health checks entirely. For instance, this can be useful if you want to perform cleanup after a complex health check or if you want to send metrics to your APM software when a health check completes. It is important to keep in mind that this hook effectively gets called on roughly the same interval you define for the health check.

## Contributing
All PR's are welcome, as long as they are well tested. Follow the typical fork->branch->pr flow.
