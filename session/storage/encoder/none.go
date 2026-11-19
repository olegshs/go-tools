package encoder

// None stores data unchanged, for binary column types.
type None struct {
}

// Encode returns src unchanged.
func (encoder *None) Encode(src []byte) ([]byte, error) {
	return src, nil
}

// Decode returns src unchanged.
func (encoder *None) Decode(src []byte) ([]byte, error) {
	return src, nil
}
