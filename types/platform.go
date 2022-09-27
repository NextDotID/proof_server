package types

type Platform string

// Platforms is a list of all current supported platforms, DO NOT MODIFY IT IN RUNTIME.
var Platforms = struct {
	Github   Platform
	NextID   Platform
	Twitter  Platform
	Keybase  Platform
	Ethereum Platform
	Discord  Platform
	Das      Platform
	Solana   Platform
	Minds    Platform
	DNS      Platform
}{
	Github:   "github",
	NextID:   "nextid",
	Twitter:  "twitter",
	Keybase:  "keybase",
	Ethereum: "ethereum",
	Discord:  "discord",
	Das:      "dotbit",
	Solana:   "solana",
	Minds:    "minds",
	DNS:      "dns",
}
