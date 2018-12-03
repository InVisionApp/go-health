# OnComplete Hook

A function can be registered with the `OnComplete` hook by providing a func to the `OnComplete` field of the `health.Config` struct when setting up your health check. The function should have the following signature: `func (state *health.State)`.
The field can be left out entirely or set to `nil` if you do not wish to register a function with the hook.

Here is a simple example of configuring a `OnComplete` hook for a health check:

```golang
func main() {
    // Create a new health instance
    h := health.New()

    // Instantiate your custom check
    cc := &customCheck{}

    // Add the checks to the health instance
    h.AddChecks([]*health.Config{
        {
            Name:     "good-check",
            Checker:  cc,
            Interval: time.Duration(2) * time.Second,
            Fatal:    true,
            OnComplete: MyHookFunc
        },
    })

    ...
}

func MyHookFunc(state *health.State) {
    log.Println("The state of %s is %s", state.Name, state.Status)
    // Other custom logic here...
}
```