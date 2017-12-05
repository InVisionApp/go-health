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

With that said, not everyone _needs_ synchronous checks. If your service has one dependency (and that is unlikely to change), it is trivial to write a basic, synchronous check and it will probably suffice.

However, if you anticipate that your service will have several dependencies, with varying degrees of complexity for determing their health state - you should probably think about introducing asynchronous health checks.

## How does this library help?
Writing an async healthchecking framework for your service is not a trivial task, especially if Go is not your primary language.

This library:

* Allows you to define how to check your dependencies.
* Allows you to define warning and fatal thresholds.
* Will run your dependency checks on a given interval, in the background.
* Exposes a way for you to gather the check results in a *fast* and *thread-safe* manner to help determine the final status of your `/status` endpoint.
* Comes bundled w/ a number of checkers for well-known dependencies such as `MySQL`, `PostgreSQL`, `Redis`, `HTTP`.
* Makes it simple to implement and provide your own checkers (by adhering to the checker interface).
* Is test-friendly
    + Provides an easy way to disable dependency health checking.
    + Uses an interface for its dependencies, allowing you to insert fakes/mocks at test time.

## Example

1. Instantiate `health` and provide it with checkers:

```go
hc := health.New()

// Create a MySQL checker
mysqlChecker, err := checkers.NewMySQL(
    &checkers.MySQL{
        User: mysqlUsername,
        Password: mysqlPassword,
        Host: mysqlHost,
        Port: mysqlPort,
        DB: mysqlDB,
        Query: `SELECT id FROM table LIMIT 1`,
        QueryTimeout: time.Duration(1) * time.Second, // Optional
        Timeout: time.Duration(5) * time.Second, // Optional
    }
)
if err != nil {
    log.Fatalf("Unable to instantiate MySQL checker: %v", err)
}

if err := hc.AddCheckers([]*HealthConfig{
    {
        Name: "my-main-mysql-dep",
        Checker: mysqlChecker,
        Fatal: true // Whether the failure of this check should cause the entire healthcheck to fail
        Interval: time.Duration(3) * time.Second,
    },
}); err != nil {
    log.Fatalf("Unable to complete adding checker configs: %v", err)
}
```

2. Start the healthchecker and pass it to the API

```go
if err := hc.Start(); err != nil {
    log.Fatalf("Unable to start healthchecker: %v", err)
}

a := api.New(hc)
log.Fatal(api.ListenAndServe())
```

3. Tie the built-in status handler to a route in your API:

```
routes := mux.NewRouter().StrictSlash(true)

routes.Handle(
    "/status", http.HandlerFunc(a.hc.HandlerJSON),
).Methods("GET")

// or use `a.hc.HandlerBasic` for simple ok|error non-JSON return
```

## Additional Documentation
* API documentation
* Bundled checker documentation
* Examples

## Contributing
All PR's are welcome, as long as they are well tested. Follow the typical fork->branch->pr flow.
