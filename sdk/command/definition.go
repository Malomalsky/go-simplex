package command

type NetworkUsage string

const (
	NetworkUsageNo          NetworkUsage = "no"
	NetworkUsageInteractive NetworkUsage = "interactive"
	NetworkUsageBackground  NetworkUsage = "background"
	NetworkUsageUnknown     NetworkUsage = "unknown"
)

type Parameter struct {
	Name string
	Type string
}

type Definition struct {
	Name         string
	Category     string
	Description  string
	NetworkUsage NetworkUsage
	Syntax       string
	Parameters   []Parameter
}

