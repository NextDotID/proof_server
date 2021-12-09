package types

type Action string

var Actions = struct {
	Create Action
	Delete Action
}{
	Create: "create",
	Delete: "delete",
}
