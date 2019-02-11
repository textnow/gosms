package gosms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test maximizes message size for a single SMS (140 bytes) and ensures that the message is not split
func TestCreateSMSPayloadsReturnsSingleSMS(t *testing.T) {
	const from = "from"
	const to = "to"

	var TestCreateSMSPayloadsReturnsSingleSMS = []struct {
		name             string
		from             string
		to               []string
		message          string
		encoder          Encoder
		shortReference   bool
		expectedSplit    []string
		expectedEncoding string
	}{
		{
			"7-bit with no special characters",
			from,
			[]string{to},
			"This message is exactly 160 characters long ....................................................................................................................",
			NewGSM(),
			true,
			[]string{"This message is exactly 160 characters long ...................................................................................................................."},
			EncoderNameGSM,
		},
		{
			"7-bit with extended GSM characters",
			from,
			[]string{to},
			"This message contains 98 regular characters and 31 special characters, totalling 160 code points. [][][][][][][][][][][][][][][]~",
			NewGSM(),
			true,
			[]string{"This message contains 98 regular characters and 31 special characters, totalling 160 code points. [][][][][][][][][][][][][][][]~"},
			EncoderNameGSM,
		},
		{
			"Unicode with only single code point characters",
			from,
			[]string{to},
			"This message contains 70 single code point characters. 你好朋友你好朋友你好朋友你好朋",
			NewUTF16(),
			true,
			[]string{"This message contains 70 single code point characters. 你好朋友你好朋友你好朋友你好朋"},
			EncoderNameUTF16,
		},
		{
			"Unicode with some double code point characters",
			from,
			[]string{to},
			"This message has 64 normal characters and 3 special characters. 🙃🙃🙃", // total 70 code points
			NewUTF16(),
			true,
			[]string{"This message has 64 normal characters and 3 special characters. 🙃🙃🙃"},
			EncoderNameUTF16,
		},
	}

	for _, tt := range TestCreateSMSPayloadsReturnsSingleSMS {
		SMSs := CreateSMSPayloads(tt.from, tt.to, tt.message, tt.encoder, tt.shortReference)

		// check SMS field correctness
		assert.Equal(t, 1, len(SMSs))
		assert.Equal(t, from, SMSs[0].from)
		assert.Equal(t, to, SMSs[0].to)
		assert.Equal(t, tt.message, SMSs[0].content)
		assert.Equal(t, tt.expectedEncoding, SMSs[0].encoder)
		assert.Equal(t, "", SMSs[0].udh)

		// check expected split vs actual split
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		assert.Equal(t, tt.expectedSplit[0], SMSs[0].content)
	}
}

// this test pushes just past the 140 byte SMS message limit and causes splitting
func TestCreateSMSPayloadsReturnsTwoSMSs(t *testing.T) {
	const from = "from"
	const to = "to"

	var TestCreateSMSPayloadsReturnsTwoSMSs = []struct {
		name             string
		from             string
		to               []string
		message          string
		encoder          Encoder
		shortReference   bool
		expectedSplit    []string
		expectedEncoding string
	}{
		{
			"7-bit with no special characters",
			from,
			[]string{to},
			"This message is exactly 161 characters long xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			NewGSM(),
			true,
			[]string{
				"This message is exactly 161 characters long ",
				"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
			EncoderNameGSM,
		},
		{
			"7-bit with extended GSM characters",
			from,
			[]string{to},
			"This message contains 155 regular characters and 3 special characters, totalling 161 code points. [][ xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			NewGSM(),
			true,
			[]string{
				"This message contains 155 regular characters and 3 special characters, totalling 161 code points. [][ ",
				"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
			EncoderNameGSM,
		},
		{
			"Unicode with only single code point characters",
			from,
			[]string{to},
			"This message contains 71 single code point characters. 你好朋友你好朋友你好朋友你好朋友",
			NewUTF16(),
			true,
			[]string{
				"This message contains 71 single code point characters. ",
				"你好朋友你好朋友你好朋友你好朋友",
			},
			EncoderNameUTF16,
		},
		{
			"Unicode with some double code point characters",
			from,
			[]string{to},
			"This message has 65 normal characters and 3 special characters.  🙃🙃🙃", // total 71 code points
			NewUTF16(),
			true,
			[]string{
				"This message has 65 normal characters and 3 special characters.  ",
				"🙃🙃🙃",
			},
			EncoderNameUTF16,
		},
		{
			"Unicode, no special char, no valid split points",
			from,
			[]string{to},
			"你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋", // total 71 code points
			NewUTF16(),
			true,
			[]string{
				"你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋",
				"友你好朋",
			},
			EncoderNameUTF16,
		},
	}

	for _, tt := range TestCreateSMSPayloadsReturnsTwoSMSs {
		SMSs := CreateSMSPayloads(tt.from, tt.to, tt.message, tt.encoder, tt.shortReference)

		// check SMS number
		assert.Equal(t, 2, len(SMSs))

		// check message validity
		assert.Equal(t, tt.message, SMSs[0].content+SMSs[1].content)

		// check UDH validity
		udh0 := []byte(SMSs[0].udh)
		udh1 := []byte(SMSs[1].udh)
		assert.NotEqual(t, 0, len(udh0))
		assert.Equal(t, len(udh0), len(udh1))
		for idx := 0; idx < len(udh0)-1; idx++ {
			assert.Equal(t, udh0[idx], udh1[idx])
		}
		assert.Equal(t, udh0[len(udh0)-1]+1, udh1[len(udh1)-1])

		// check SMS field correctness
		assert.Equal(t, from, SMSs[0].from)
		assert.Equal(t, to, SMSs[0].to)
		assert.Equal(t, tt.expectedEncoding, SMSs[0].encoder)
		assert.Equal(t, from, SMSs[1].from)
		assert.Equal(t, to, SMSs[1].to)
		assert.Equal(t, tt.expectedEncoding, SMSs[1].encoder)

		// check expected split vs actual split
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		for idx := range SMSs {
			assert.Equal(t, tt.expectedSplit[idx], SMSs[idx].content)
		}
	}
}

// this test passes CreateSMSPayloads multiple recipients and ensures that each resulting
// SMS has recipients listed correctly
func TestCreateSMSPayloadsConcatenatesTo(t *testing.T) {
	const from = "from"
	var to = []string{"to1", "to2"}
	const expectedTo = "to1 to2"
	const message = "message"
	const expectedSplitString = "message"

	var TestCreateSMSPayloadsReturnsSingleSMS = []struct {
		name             string
		from             string
		to               []string
		message          string
		encoder          Encoder
		shortReference   bool
		expectedSplit    []string
		expectedEncoding string
	}{
		{
			"small message, no splitting, two to's",
			from,
			to,
			message,
			NewGSM(),
			true,
			[]string{expectedSplitString},
			EncoderNameGSM,
		},
	}

	for _, tt := range TestCreateSMSPayloadsReturnsSingleSMS {
		SMSs := CreateSMSPayloads(tt.from, tt.to, tt.message, tt.encoder, tt.shortReference)

		// check SMS field correctness
		assert.Equal(t, 1, len(SMSs))
		assert.Equal(t, from, SMSs[0].from)
		assert.Equal(t, expectedTo, SMSs[0].to)
		assert.Equal(t, tt.message, SMSs[0].content)
		assert.Equal(t, tt.expectedEncoding, SMSs[0].encoder)
		assert.Equal(t, "", SMSs[0].udh)

		// check expected split vs actual split
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		assert.Equal(t, tt.expectedSplit[0], SMSs[0].content)
	}
}

// this test makes sure that appendUDHs does not append a UDH to SMSs if too few are given
func TestAppendUDHsMakesNoChanges(t *testing.T) {
	const from = "from"
	const to = "to"
	const content = "content"
	const encoderName = "encoderName"

	var TestAppendUDHsMakesNoChanges = []struct {
		name string
		SMSs []SMS
	}{
		{
			"no SMSs",
			nil,
		},
		{
			"one SMS",
			[]SMS{newSMS(from, to, content, encoderName, "")},
		},
	}

	for _, tt := range TestAppendUDHsMakesNoChanges {
		SMSs := appendUDHs(tt.SMSs, true)

		// check SMS number
		assert.Equal(t, len(tt.SMSs), len(SMSs))

		// check SMS content
		for _, sms := range SMSs {
			assert.Equal(t, from, sms.from)
			assert.Equal(t, to, sms.to)
			assert.Equal(t, content, sms.content)
			assert.Equal(t, encoderName, sms.encoder)
			assert.Equal(t, "", sms.udh)
		}
	}
}

// this test makes sure that appendUDHs adds a UDH correctly
func TestAppendUDHsAddsUDHWithShortReferenceNumber(t *testing.T) {
	const from = "from"
	const to = "to"
	const content = "content"
	const encoderName = "encoderName"

	var TestAppendUDHsAddsUDHWithShortReferenceNumber = []struct {
		name string
		SMSs []SMS
	}{
		{
			"just enough SMSs",
			[]SMS{
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
			},
		},
		{
			"loads of SMSs",
			[]SMS{
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
			},
		},
	}

	for _, tt := range TestAppendUDHsAddsUDHWithShortReferenceNumber {
		SMSs := appendUDHs(tt.SMSs, true)

		// check SMS number
		assert.Equal(t, len(tt.SMSs), len(SMSs))

		// check SMS content
		for _, sms := range SMSs {
			assert.Equal(t, from, sms.from)
			assert.Equal(t, to, sms.to)
			assert.Equal(t, content, sms.content)
			assert.Equal(t, encoderName, sms.encoder)
			assert.NotEqual(t, "", sms.udh)
		}

		// check UDH validity
		for idx, sms := range SMSs {
			assert.Equal(t, udhByteLengthShort, len(sms.udh))
			assert.Equal(t, int8(udhByteLengthShort-1), int8(sms.udh[0]))
			assert.Equal(t, int8(shortReferenceInfoElementID), int8(sms.udh[1]))
			assert.Equal(t, int8(udhByteLengthShort-3), int8(sms.udh[2]))
			assert.Equal(t, int8(SMSs[0].udh[3]), int8(sms.udh[3])) // all SMSs should have the same reference number
			assert.Equal(t, int8(len(SMSs)), int8(sms.udh[4]))
			assert.Equal(t, int8(idx+1), int8(sms.udh[5]))
		}
	}
}

// this test makes sure that appendUDHs adds a UDH correctly
func TestAppendUDHsAddsUDHWithLongReferenceNumber(t *testing.T) {
	const from = "from"
	const to = "to"
	const content = "content"
	const encoderName = "encoderName"

	var TestAppendUDHsAddsUDHWithLongReferenceNumber = []struct {
		name string
		SMSs []SMS
	}{
		{
			"just enough SMSs",
			[]SMS{
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
			},
		},
		{
			"loads of SMSs",
			[]SMS{
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
				newSMS(from, to, content, encoderName, ""),
			},
		},
	}

	for _, tt := range TestAppendUDHsAddsUDHWithLongReferenceNumber {
		SMSs := appendUDHs(tt.SMSs, false)

		// check SMS number
		assert.Equal(t, len(tt.SMSs), len(SMSs))

		// check SMS content
		for _, sms := range SMSs {
			assert.Equal(t, from, sms.from)
			assert.Equal(t, to, sms.to)
			assert.Equal(t, content, sms.content)
			assert.Equal(t, encoderName, sms.encoder)
			assert.NotEqual(t, "", sms.udh)
		}

		// check UDH validity
		for idx, sms := range SMSs {
			assert.Equal(t, udhByteLengthLong, len(sms.udh))
			assert.Equal(t, int8(udhByteLengthLong-1), int8(sms.udh[0]))
			assert.Equal(t, int8(longReferenceInfoElementID), int8(sms.udh[1]))
			assert.Equal(t, int8(udhByteLengthLong-3), int8(sms.udh[2]))
			assert.Equal(t, int8(SMSs[0].udh[3]), int8(sms.udh[3])) // all SMSs should have the same reference number
			assert.Equal(t, int8(SMSs[0].udh[4]), int8(sms.udh[4])) // all SMSs should have the same reference number
			assert.Equal(t, int8(len(SMSs)), int8(sms.udh[5]))
			assert.Equal(t, int8(idx+1), int8(sms.udh[6]))
		}
	}
}
