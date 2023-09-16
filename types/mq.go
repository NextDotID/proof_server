package types

type QueueAction string

var QueueActions = struct {
	Revalidate               QueueAction
	ArweaveUpload            QueueAction
	TwitterOAuthTokenAcquire QueueAction
}{
	Revalidate:               "revalidate",
	ArweaveUpload:            "arweave_upload",
	TwitterOAuthTokenAcquire: "twitter_oauth_token_acquire",
}

// QueueMessage indicates structure of messages in Amazon SQS.
type QueueMessage struct {
	Action QueueAction `json:"action"`
	// For revalidate.
	ProofID int64 `json:"proof_id"`
	// For Arweave upload.
	Persona string `json:"persona"`
}
