package spec

type Health struct {
	globalFailed bool
}

func New() *Health {
	return &Health{}
}

func (h *Health) Failed() bool {
	return true
}

type HealthcheckStatus struct {
	Name string
	Failed bool
	CauseFatal bool
	Data interface{} // contains JSON message (that can be marshalled)
}

// allow folks to use either the default handler or a custom one
hc.Handler
or
hc.HandlerFunc = hc.HandlerBasic
or
hc.HandlerFunc = mySpecialHandler

// add checks - one at a time or all at once
hcInstance.AddCheck("name", ICheckable, true||false)
hc.AddChecks(map[string]ICheckable)
if err := hc.Start(); err != nil {
...
}

// optional -- get your statuses, do something with them
statuses, failed := hc.Status() // returans map of all cached statuses []*Status
...

// optional -- use this if you want to generate your own
smi, failed := hc.StatusMapInterface() // map[string]interface{}
smi["version"] = "..."
data, err := json.Marshal(smi)

// optional
route = a.Deps.hc.Handler() // JSON
route = a.Deps.hc.HandlerBasic() // "ok" || "oops"

// optional -- let's not do this; no need
["version"] = 123
jsonBlob := StatusJSON(map[string]interface{})
