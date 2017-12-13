[![Build Status](https://travis-ci.com/InVisionApp/go-health.svg?token=KosA43m1X3ikri8JEukQ&branch=master)](https://travis-ci.com/InVisionApp/go-health)

<img align="right" src="images/go-health.svg" width="200">

# go-health
A library that enables *async* dependency health checking for services running on an orchastrated container platform such as kubernetes or mesos.

## Why is this important?
Container orchestration platforms require that the underlying service(s) expose a "healthcheck" which is used by the platform to determine whether the container is in a good or bad state.

While this can be achieved by simply exposing a `/status` endpoint that perfoms synchronous checks against its dependencies (followed by returning a `200` or `non-200` status code), it is not optimal for a number of reasons:

* **It does not scale**
    + The more dependencies you add, the longer your healthcheck will take to complete (and potentially cause your service to be killed off by the orchestration platform).
    + Depending on the complexity of a given dependency, your check may be fairly involved where it is _okay_ for it to take `30s+` to complete.
* **It adds unnecessary load on yours deps or at worst, becomes a DoS target**
    + **Non-malicious scenario**
        + Thundering herd problem -- in the event of a deployment (or restart, etc.), all of your service containers are likely to have their `/status` endpoints checked by the orchestration platform as soon as they come up. Depending on the complexity of the checks, running that many simultaneous checks against your dependencies could cause at worst the dependencies to experience problems and at minimum add unnecessary load.
        + Security scanners -- if your organization runs periodic security scans, they may hit your `/status` endpoint and trigger unnecessary dep checks.
    + **Malicious scenario**
        + Loading up any basic HTTP benchmarking tool and pointing it at your `/status` endpoint could choke your dependencies (and potentially your service).

With that said, not everyone _needs_ asynchronous checks. If your service has one dependency (and that is unlikely to change), it is trivial to write a basic, synchronous check and it will probably suffice.

However, if you anticipate that your service will have several dependencies, with varying degrees of complexity for determing their health state - you should probably think about introducing asynchronous health checks.

## How does this library help?
Writing an async healthchecking framework for your service is not a trivial task, especially if Go is not your primary language.

This library:

* Allows you to define how to check your dependencies.
* Allows you to define warning and fatal thresholds.
* Will run your dependency checks on a given interval, in the background.
* Exposes a way for you to gather the check results in a *fast* and *thread-safe* manner to help determine the final status of your `/status` endpoint.
* Comes bundled w/ a number of checkers for well-known dependencies such as `MySQL`, `PostgreSQL`, `Redis`, `HTTP`, `Mongo`.
* Makes it simple to implement and provide your own checkers (by adhering to the checker interface).
* Is test-friendly
    + Provides an easy way to disable dependency health checking.
    + Uses an interface for its dependencies, allowing you to insert fakes/mocks at test time.

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
)
```

3. Start the healthcheck

```golang
h.Start()
```

From here on, you can either configure an endpoint such as `/healthcheck` to use a built-in handler such as `handlers.NewJSONHandlerFunc()` or get the current health state of all your deps by traversing the data returned by `h.State()`.

## Additional Documentation
* [Examples](/examples)
* [Checkers](/checkers)
* [Logging](/loggers)

## Contributing
All PR's are welcome, as long as they are well tested. Follow the typical fork->branch->pr flow.
