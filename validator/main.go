package validator

type IValidator interface {
	// GenerateSignPayload generates a string to be signed.
	GenerateSignPayload() (payload string)

	Validate() (result bool)
}
