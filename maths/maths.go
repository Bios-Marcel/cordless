package maths

// Min returns the smaller of the passed numbers.1
func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

// Max returns the bigger of the passed numbers.
func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
