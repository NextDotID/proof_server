package types

type Action string

var Actions = struct {
	Create Action
	Delete Action
	KV  Action
}{
	Create: "create",
	Delete: "delete",
	KV:  "kv",
}
