/*
 * Copyright 2023 FormulaGo Authors
 *
 * Created by hua
 */

package types

// SubStrByLen UTF-8 is used as the standard for character truncation
// speed test, 57.80 ns/op
func SubStrByLen(text string, length int) string {
	// n is the character position for each traversal, and i is the number of bytes per traversal
	// (depending on how many bytes each character occupies)
	var n, i int
	// I iterate through the text starting from 0,
	// and the number of bytes of the corresponding character is stacked each time
	for i = range text {
		n++
		if n > length {
			break
		}
	}
	// If n is less than or equal to length, the original string is returned
	if n <= length {
		return text
	}
	return text[:i]
}
