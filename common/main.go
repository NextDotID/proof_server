package common

type Runtime string

var Runtimes = struct {
	Standalone Runtime
	Lambda     Runtime
}{
	Standalone: "standalone",
	Lambda:     "lambda",
}

var (
	CurrentRuntime = Runtimes.Standalone
	Environment    = "unknown"
	Revision       = "UNKNOWN"
	BuildTime      = "0"
)
