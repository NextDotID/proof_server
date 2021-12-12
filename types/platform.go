package types

type Platform string

// Platforms is a list of all current supported platforms, DO NOT MODIFY IT IN RUNTIME.
var Platforms = struct {
	NextID   Platform
	Twitter  Platform
	Keybase  Platform
	Ethereum Platform
}{
	NextID: "nextid",
	Twitter:  "twitter",
	Keybase:  "keybase",
	Ethereum: "ethereum",
}
