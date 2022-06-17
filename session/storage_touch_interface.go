package session

type StorageTouchInterface interface {
	Touch(string) error
}
