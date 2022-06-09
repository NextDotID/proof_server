package types

type QueueAction string

var QueueActions = struct {
	Revalidate    QueueAction
	ArweaveUpload QueueAction
}{
	Revalidate:    "revalidate",
	ArweaveUpload: "arweave_upload",
}

// QueueMessage indicates structure of messages in Amazon SQS.
type QueueMessage struct {
	Action  QueueAction `json:"action"`
	ProofID int64       `json:"proof_id"`
}
