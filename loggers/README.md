loggers
===
The `loggers` package provides you with a way to utilize either a pre-built logger shim for popular loggers or write a custom logger shim that implements the `ILogger` interface (and thus is usable by the `health` lib).

This has to be done since the `loggers` package in the standard library does not provide a logger interface.

## Options
By default, `health` will utilize the standard library `loggers` package.

If you do not wish for `health` to perform any sort of logging, you can update `h.Logger` to point to a noop logger: `h.Logger = loggers.NewNoopLogger()`.

## Example w/ logrus
```golang
import (
    "github.com/InVisionApp/go-health"
    "github.com/InVisionApp/go-health/loggers"
    "github.com/InVisionApp/go-health/checkers"
)

// create and configure a health instance
h := health.New()
h.AddChecks(...)

// Set the logger
h.Logger = loggers.NewLogrus(nil)

// Or alternatively, you can provide your own logrus instance
myLogrus := logrus.WithField("foo", "bar")
h.Logger = loggers.NewLogrus(myLogrus)

// Start healthcheck
h.Start()
```

## Example w/ Noop logger
```golang
import (
    "github.com/InVisionApp/go-health"
    "github.com/InVisionApp/go-health/loggers"
    "github.com/InVisionApp/go-health/checkers"
)

// create and configure a health instance
h := health.New()
h.AddChecks(...)

h.Logger = loggers.NewNoopLogger()

// Start healthcheck
h.Start()
```
