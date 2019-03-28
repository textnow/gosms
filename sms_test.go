package gosms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test ensures that the fields of SMS are set and fetched correctly
func TestSMS(t *testing.T) {
	const from = "from"
	const to = "to"
	const content = "content"
	const udh = "udh"

	sms := newSMS(from, to, content, udh)

	assert.Equal(t, from, sms.GetFrom())
	assert.Equal(t, to, sms.GetTo())
	assert.Equal(t, content, sms.GetContent())
	assert.Equal(t, udh, sms.GetUDH())
}
