
package gosms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test ensures that all GSM related functions work as expected
func TestGSM(t *testing.T) {
	encoder := NewGSM()

	// check that the right constants are returned from GetCodePointBits and GetEncoderName
	assert.Equal(t, codePointBitsGSM, encoder.GetCodePointBits())
	assert.Equal(t, EncoderNameGSM, encoder.GetEncoderName())

	// check that code point are reported accurately for each GSM character
	for char, expectedCodePoints := range gsmCodePoints {
		actualCodePoints, err := encoder.GetCodePoints(char)

		assert.Equal(t, expectedCodePoints, actualCodePoints)
		assert.Nil(t, err)
	}

	// check that non-GSM characters cause GetCodePoints to return an error
	expectedError := ErrNotEncodable
	codePoints, err := encoder.GetCodePoints('你')

	assert.Zero(t, codePoints)
	assert.EqualError(t, expectedError, err.Error())

	// check that all GSM characters cause CheckEncodability to return true
	for char, _ := range gsmCodePoints {
		encodable := encoder.CheckEncodability(string(char))
		assert.True(t, encodable)
	}

	// check that non-GSM characters cause CheckEncodability to return false
	encodable := encoder.CheckEncodability("你")
	assert.False(t, encodable)
}

// this test ensures that all UTF16 related functions work as expected
func TestUTF16(t *testing.T) {
	encoder := NewUTF16()

	// check that the right constants are returned from GetCodePointBits and GetEncoderName
	assert.Equal(t, codePointBitsUTF16, encoder.GetCodePointBits())
	assert.Equal(t, EncoderNameUTF16, encoder.GetEncoderName())

	// check that code point are reported accurately for UTF16 characters
	codePoints, err := encoder.GetCodePoints('😂') // 2 code points
	assert.Equal(t, codePoints, 2)
	assert.Nil(t, err)

	codePoints, err = encoder.GetCodePoints('#') // 1 code point
	assert.Equal(t, codePoints, 1)
	assert.Nil(t, err)

	// all golang strings should cause CheckEncodability to return true since they contain unicode characters
	encodable := encoder.CheckEncodability("你")
	assert.True(t, encodable)
}
