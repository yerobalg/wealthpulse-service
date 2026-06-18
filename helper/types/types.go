package types

// SafelyDereference is a utility function that safely dereferences an interface.
//
// Parameters:
//   - input: The interface to be dereferenced.
//
// Returns:
//   - The dereferenced interface.
func SafelyDereference[T any](input *T) T {
	if input == nil {
		var data T
		return data
	}
	return *input
}

// SafelyReference is a utility function that safely references an interface.
//
// Parameters:
//   - input: The interface to be referenced.
//
// Returns:
//   - The referenced interface.
func SafelyReference[T any](input T) *T {
	return &input
}
