package gosms

import (
	"errors"
	"unicode"
)

// ErrNotSplittable indicates that the given message cannot be split with the given encoder and message length
var ErrNotSplittable = errors.New("the message cannot be split with the given encoder and message length")

// willMessageFit checks to see it a message will fit in a messageLength space without being split
func willMessageFit(message []rune, encoder Encoder, messageLength int) (bool, error) {
	var codePoints int

	for _, char := range message {
		// Some encodings have variable lengthed characters
		charPoints, err := encoder.GetCodePoints(char)
		if err != nil {
			return false, ErrNotEncodable
		}

		codePoints += charPoints

		if codePoints > messageLength {
			return false, nil
		}
	}
	return true, nil
}

// Returns true if it is safe to split a message before char
func canSplitBefore(char rune) bool {
	// Numbers, graphics and words should not be split if possible.
	// Marks include accents which should stay connected to letters
	// Symbols include emojis, which we should avoid splitting if possible

	// Control Characters are not a part of words, numbers or graphics.
	// Spaces are not a part of words, numbers. They are usually used to delimit words
	// Punctuation is not a part of words or numbers, however it should be appended
	//  to whatever words or numbers precede it
	return unicode.IsControl(char) || unicode.IsSpace(char)
}

// Returns true if it is safe to split a message after char
func canSplitAfter(char rune) bool {
	// Numbers, graphics and words should not be split if possible.
	// Marks include accents which should stay connected to letters
	// Symbols include emojis, which we should avoid splitting if possible

	// Control Characters are not a part of words, numbers or graphics.
	// Spaces are not a part of words, numbers. They are usually used to delimit words
	// Punctuation is not a part of words or numbers
	return unicode.IsControl(char) || unicode.IsSpace(char) || unicode.IsPunct(char)
}

// SplitMessage splits a message into parts with a maximum length of messageLength
// code points. Word splitting is avoided.
func SplitMessage(message []rune, encoder Encoder, messageLength int) ([]string, error) {
	var messageParts []string
	var messagePart []rune
	var codePoints int
	var lastSplitPoint = -1 // no valid split point

	for idx := 0; idx < len(message); idx++ {
		var char = message[idx]

		// Some encodings have variable lengthed characters
		charPoints, err := encoder.GetCodePoints(char)
		if err != nil {
			return nil, ErrNotEncodable
		}

		codePoints += charPoints

		// check for split point
		if canSplitBefore(char) {
			lastSplitPoint = len(messagePart)
		}

		// if the SMS is full
		if codePoints > messageLength {
			// if the split is impossible
			if len(messagePart) == 0 {
				return nil, ErrNotSplittable
			}

			// split at the last valid point
			if lastSplitPoint == -1 {
				lastSplitPoint = len(messagePart)
			}

			// recover dropped characters
			idx -= (len(messagePart) - lastSplitPoint)
			messagePart = messagePart[0:lastSplitPoint]

			// save message part
			messageParts = append(messageParts, string(messagePart))

			// reset
			messagePart = nil
			codePoints = 0
			lastSplitPoint = -1

			// try adding char again with fresh message part
			idx--
			continue
		}

		// add char to message part
		messagePart = append(messagePart, char)

		// check for split point
		if canSplitAfter(char) {
			lastSplitPoint = len(messagePart)
		}
	}

	// save last message part
	messageParts = append(messageParts, string(messagePart))

	return messageParts, nil
}
