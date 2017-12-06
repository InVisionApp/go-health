log
===
The `log` package provides you with a way to utilize either a pre-built logger shim for popular loggers or write a custom logger shim that implements the `ILogger` interface (and thus is usable by the `health` lib).

This has to be done since the `log` package in the standard library does not provide a logger interface.

## Options
By default, `health` will utilize the standard library `log` package.

If you do not wish to perform any logging, you can update the `h.Logger` to point to a noop logger: `h.Logger = log.NewNoopLogger()`.

## Example w/ logrus
```golang
// create and configure a health instance
h := health.New()
h.AddChecks(...)

// Set the logger
h.Logger = log.NewLoggerLogrus(nil)

// Or alternatively, you can provide your own logrus instance
myLogrus := logrus.WithField("foo", "bar")
h.Logger = log.NewLoggerLogrus(myLogrus)

// Start healthcheck
h.Start()
```