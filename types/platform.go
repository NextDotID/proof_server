package types

type Platform string

// Platforms is a list of all current supported platforms, DO NOT MODIFY IT IN RUNTIME.
var Platforms = struct {
	Github      Platform
	NextID      Platform
	Twitter     Platform
	Telegram    Platform
	Keybase     Platform
	Ethereum    Platform
	Discord     Platform
	Das         Platform
	Solana      Platform
	Minds       Platform
	DNS         Platform
	ENS         Platform
	Steam       Platform
	ActivityPub Platform
	Slack       Platform
}{
	Github:      "github",
	NextID:      "nextid",
	Twitter:     "twitter",
	Telegram:    "telegram",
	Keybase:     "keybase",
	Ethereum:    "ethereum",
	Discord:     "discord",
	Das:         "dotbit",
	Solana:      "solana",
	Minds:       "minds",
	DNS:         "dns",
	ENS:         "ens",
	Steam:       "steam",
	ActivityPub: "activitypub",
	Slack:       "slack",
}
