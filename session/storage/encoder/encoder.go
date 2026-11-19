// Package encoder converts session data to and from the form stored in the
// database by SqlStorage.
package encoder

const (
	// Encoder names.
	EncoderBase64 = "base64"
	EncoderNone   = "none"
)

// Encoder converts session data for storage and back.
type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

// New returns the encoder with the given name. Unknown names get base64.
func New(name string) Encoder {
	switch name {
	default:
		return new(Base64)
	case EncoderBase64:
		return new(Base64)
	case EncoderNone:
		return new(None)
	}
}
