package interfaces

// Helper hides the differences between SQL dialects.
type Helper interface {
	// ArgPlaceholder returns the placeholder for the i-th argument (from 1).
	ArgPlaceholder(int) string
	// EscapeName quotes a table or column name, which may be qualified
	// ("t.id") or have an alias ("users u").
	EscapeName(string) string
	// IsDuplicateKey reports whether err is a unique constraint violation.
	IsDuplicateKey(error) bool
}
