package session

// StorageTouchInterface is implemented by storages that can refresh a
// session's expiry time without rewriting its data.
type StorageTouchInterface interface {
	Touch(string) error
}
