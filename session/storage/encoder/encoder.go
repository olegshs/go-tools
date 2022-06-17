package encoder

const (
	EncoderBase64 = "base64"
	EncoderNone   = "none"
)

type Encoder interface {
	Encode([]byte) ([]byte, error)
	Decode([]byte) ([]byte, error)
}

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
