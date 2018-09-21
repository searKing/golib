package object

import "github.com/searKing/golib/util/preconditions"

// Checks if the {@code index} is within the bounds of the range from
// {@code 0} (inclusive) to {@code length} (exclusive).
func CheckIndex(index, length int) int {
	return preconditions.CheckIndex(index, length)
}

// Checks if the sub-range from {@code fromIndex} (inclusive) to
// {@code toIndex} (exclusive) is within the bounds of range from {@code 0}
// (inclusive) to {@code length} (exclusive).
func CheckFromToIndex(fromIndex, toIndex, length int) int {
	return preconditions.CheckFromToIndex(fromIndex, toIndex, length);
}

// Checks if the sub-range from {@code fromIndex} (inclusive) to
// {@code fromIndex + size} (exclusive) is within the bounds of range from
// {@code 0} (inclusive) to {@code length} (exclusive).
func CheckFromIndexSize(fromIndex, size, length int) int {
	return preconditions.CheckFromIndexSize(fromIndex, size, length);
}
