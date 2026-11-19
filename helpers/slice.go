package helpers

// Slice is a slice with helper methods for comparable element types.
type Slice[T comparable] []T

// IndexOf returns the index of the first element equal to value, or -1 if
// there is none.
func (slice Slice[T]) IndexOf(value T) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

// Map returns a new slice with f applied to every element.
func (slice Slice[T]) Map(f func(T) T) Slice[T] {
	n := len(slice)
	a := make(Slice[T], n)

	for i, v := range slice {
		a[i] = f(v)
	}

	return a
}
