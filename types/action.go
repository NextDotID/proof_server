package types

type Action string

var Actions = struct {
	Create Action
	Delete Action
	KVSet  Action
}{
	Create: "create",
	Delete: "delete",
	KVSet:  "kv_set",
}
