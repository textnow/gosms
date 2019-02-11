package gosms

import "unicode/utf16"

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

// Extended GSM characters set, each counts as 2 code points
var extendedGSMChars = map[rune]string{
	12:   "\f",
	91:   "[",
	92:   "\\",
	93:   "]",
	94:   "^",
	123:  "{",
	124:  "|",
	125:  "}",
	126:  "~",
	8364: "€",
}

// Encoder structure encapsulates encoder specific fields
type Encoder interface {
	GetEncoderName() string
	GetCodePointBits() int
	GetCodePoints(rune) int
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
func (s *GSM) GetCodePoints(char rune) int {
	_, isGsm7Ext := extendedGSMChars[char]
	if isGsm7Ext {
		return 2
	}
	return 1
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
func (s *UTF16) GetCodePoints(char rune) int {
	utf16Rune, _ := utf16.EncodeRune(char)
	if utf16Rune >= highSurrogateStart && utf16Rune <= highSurrogateEnd {
		return 2
	}
	return 1
}
