handlers
========
The `health` library comes bundled with some no-thrills HTTP handlers that should
fit _most_ use cases.

## Usage
After setting up an instance of `health`, register a healthcheck endpoint and point
it at a handler func.

```golang
import (
    "github.com/InVisionApp/go-health/v2"
    "github.com/InVisionApp/go-health/v2/checkers"
    "github.com/InVisionApp/go-health/v2/handlers"
)

// create and configure a new health instance
h := health.New()
h.AddChecks(...)

// Register a new endpoint and have it use a pre-built handler
http.HandleFunc("/healthcheck", handlers.NewJSONHandlerFunc(h, nil))
http.ListenAndServe(":8080", nil)
```

## Behavior
If any check fails that is configured as `fatal` - the handler will return a
`http.StatusInternalServerError`; otherwise, it will return a `http.StatusOK`.

## `handlers.NewJSONHandlerFunc` output example
```json
{
    "details": {
        "bad-check": {
            "name": "bad-check",
            "status": "failed",
            "error": "Ran into error while performing 'GET' request: Get google.com: unsupported protocol scheme \"\"",
            "check_time": "2017-12-05T19:17:23.691637151-08:00"
        },
        "good-check": {
            "name": "good-check",
            "status": "ok",
            "check_time": "2017-12-05T19:17:23.857481271-08:00"
        }
    },
    "status": "ok"
}
```

## `handlers.NewBasicHandlerFunc` example output
```
ok || failed
```
