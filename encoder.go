package gosms

import (
	"errors"
	"unicode/utf16"
)

const (
	// EncoderNameGSM is the GSM Encoder Name
	EncoderNameGSM string = "GSM"

	// EncoderNameUTF16 is the UTF-16 Encoder Name
	EncoderNameUTF16 string = "UTF-16"

	codePointBitsGSM   int  = 7
	codePointBitsUTF16 int  = 16
	highSurrogateStart rune = 0xD800
	highSurrogateEnd   rune = 0xDBFF
)

// Encoder encapsulates encoder specific fields
type Encoder interface {
	GetEncoderName() string
	GetCodePointBits() int
	GetCodePoints(rune) (int, error)
}

// GSM implements the Encoder interface
type GSM struct{}

// NewGSM returns a new gsm
func NewGSM() Encoder {
	return &GSM{}
}

// GetCodePointBits returns the number of bits that make a single GSM code point
func (s *GSM) GetCodePointBits() int {
	return codePointBitsGSM
}

// GetEncoderName returns the GSM encoder name
func (s *GSM) GetEncoderName() string {
	return EncoderNameGSM
}

// GetCodePoints returns the number of code points used to represent char in GSM
func (s *GSM) GetCodePoints(char rune) (int, error) {
	codePoints, isGSM := gsmCodePoints[char]
	if !isGSM {
		return 0, errors.New("char does not belong to the GSM character set")
	}
	return codePoints, nil
}

// UTF16 implements the Encoder interface
type UTF16 struct{}

// NewUTF16 returns a new UTF16
func NewUTF16() Encoder {
	return &UTF16{}
}

// GetCodePointBits returns the number of bits that make a single UTF-16 code point
func (s *UTF16) GetCodePointBits() int {
	return codePointBitsUTF16
}

// GetEncoderName returns the UTF-16 encoder name
func (s *UTF16) GetEncoderName() string {
	return EncoderNameUTF16
}

// GetCodePoints returns the number of code points used to represent char in UTF-16
func (s *UTF16) GetCodePoints(char rune) (int, error) {
	utf16Rune, _ := utf16.EncodeRune(char)
	if utf16Rune >= highSurrogateStart && utf16Rune <= highSurrogateEnd {
		return 2, nil
	}
	return 1, nil
}
