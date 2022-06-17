package interfaces

type Helper interface {
	ArgPlaceholder(int) string
	EscapeName(string) string
	IsDuplicateKey(error) bool
}
