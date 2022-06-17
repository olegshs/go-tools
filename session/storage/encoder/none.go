package encoder

type None struct {
}

func (encoder *None) Encode(src []byte) ([]byte, error) {
	return src, nil
}

func (encoder *None) Decode(src []byte) ([]byte, error) {
	return src, nil
}
