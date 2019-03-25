package gosms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test maximizes message size for a single SMS (140 bytes) and ensures that the message is not split
func TestSplitReturnsSingleSMS(t *testing.T) {
	const from = "from"
	const to = "to"

	var TestSplitReturnsSingleSMS = []struct {
		name             string
		from             string
		to               []string
		message          string
		expectedSplit    []string
	}{
		{
			"7-bit with no special characters",
			from,
			[]string{to},
			"This message is exactly 160 characters long ....................................................................................................................",
			[]string{"This message is exactly 160 characters long ...................................................................................................................."},
		},
		{
			"7-bit with extended GSM characters",
			from,
			[]string{to},
			"This message contains 98 regular characters and 31 special characters, totalling 160 code points. [][][][][][][][][][][][][][][]~",
			[]string{"This message contains 98 regular characters and 31 special characters, totalling 160 code points. [][][][][][][][][][][][][][][]~"},
		},
		{
			"Unicode with only single code point characters",
			from,
			[]string{to},
			"This message contains 70 single code point characters. 你好朋友你好朋友你好朋友你好朋",
			[]string{"This message contains 70 single code point characters. 你好朋友你好朋友你好朋友你好朋"},
		},
		{
			"Unicode with some double code point characters",
			from,
			[]string{to},
			"This message has 64 normal characters and 3 special characters. 🙃🙃🙃", // total 70 code points
			[]string{"This message has 64 normal characters and 3 special characters. 🙃🙃🙃"},
		},
	}

	for _, tt := range TestSplitReturnsSingleSMS {
		// use default settings and auto-detect encoder
		splitter := NewSplitter()
		SMSs, err := splitter.Split(tt.from, tt.to, tt.message)
		if err != nil {
			t.Fatalf("an error '%s' was encountered when splitting the message for test '%s'", err, tt.name)
		}

		// check SMS field correctness
		assert.Equal(t, 1, len(SMSs))
		assert.Equal(t, from, SMSs[0].from)
		assert.Equal(t, to, SMSs[0].to)
		assert.Equal(t, tt.message, SMSs[0].content)
		assert.Equal(t, "", SMSs[0].udh)

		// check expected split vs actual split
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		assert.Equal(t, tt.expectedSplit[0], SMSs[0].content)
	}
}

// this test pushes just past the 140 byte SMS message limit and causes splitting
func TestSplitReturnsTwoSMSs(t *testing.T) {
	const from = "from"
	const to = "to"

	var TestSplitReturnsTwoSMSs = []struct {
		name             string
		from             string
		to               []string
		message          string
		expectedSplit    []string
	}{
		{
			"7-bit with no special characters",
			from,
			[]string{to},
			"This message is exactly 161 characters long xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			[]string{
				"This message is exactly 161 characters long ",
				"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
		},
		{
			"7-bit with extended GSM characters",
			from,
			[]string{to},
			"This message contains 155 regular characters and 3 special characters, totalling 161 code points. [][ xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			[]string{
				"This message contains 155 regular characters and 3 special characters, totalling 161 code points. [][ ",
				"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			},
		},
		{
			"Unicode with only single code point characters",
			from,
			[]string{to},
			"This message contains 71 single code point characters. 你好朋友你好朋友你好朋友你好朋友",
			[]string{
				"This message contains 71 single code point characters. ",
				"你好朋友你好朋友你好朋友你好朋友",
			},
		},
		{
			"Unicode with some double code point characters",
			from,
			[]string{to},
			"This message has 65 normal characters and 3 special characters.  🙃🙃🙃", // total 71 code points
			[]string{
				"This message has 65 normal characters and 3 special characters.  ",
				"🙃🙃🙃",
			},
		},
		{
			"Unicode, no special char, no valid split points",
			from,
			[]string{to},
			"你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋", // total 71 code points
			[]string{
				"你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋友你好朋",
				"友你好朋",
			},
		},
	}

	for _, tt := range TestSplitReturnsTwoSMSs {
		// use default settings and auto-detect encoder
		splitter := NewSplitter()
		SMSs, err := splitter.Split(tt.from, tt.to, tt.message)
		if err != nil {
			t.Fatalf("an error '%s' was encountered when splitting the message for test '%s'", err, tt.name)
		}

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
		assert.Equal(t, from, SMSs[1].from)
		assert.Equal(t, to, SMSs[1].to)

		// check expected split vs actual split
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		for idx := range SMSs {
			assert.Equal(t, tt.expectedSplit[idx], SMSs[idx].content)
		}
	}
}

// this test passes Split multiple recipients and ensures that each resulting
// SMS has recipients listed correctly
func TestSplitConcatenatesTo(t *testing.T) {
	const from = "from"
	var to = []string{"to1", "to2"}
	const expectedTo = "to1 to2"
	const message = "message"
	const expectedSplitString = "message"

	var TestSplitConcatenatesTo = []struct {
		name             string
		from             string
		to               []string
		message          string
		expectedSplit    []string
	}{
		{
			"small message, no splitting, two to's",
			from,
			to,
			message,
			[]string{expectedSplitString},
		},
	}

	for _, tt := range TestSplitConcatenatesTo {
		// use default settings and auto-detect encoder
		splitter := NewSplitter()
		SMSs, err := splitter.Split(tt.from, tt.to, tt.message)
		if err != nil {
			t.Fatalf("an error '%s' was encountered when splitting the message for test '%s'", err, tt.name)
		}

		// check SMS field correctness
		assert.Equal(t, 1, len(SMSs))
		assert.Equal(t, from, SMSs[0].from)
		assert.Equal(t, expectedTo, SMSs[0].to)
		assert.Equal(t, tt.message, SMSs[0].content)
		assert.Equal(t, "", SMSs[0].udh)

		// check expected split vs actual split
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		assert.Equal(t, tt.expectedSplit[0], SMSs[0].content)
	}
}

// this test ensures that the Splitter's configuration setter functions work as expected
func TestSplitUsesSpecifiedConfiguration(t *testing.T) {
	const from = "from"
	var to = "to"
	const message = "All of the characters that make up this message are in the GSM character set." // 77 code points

	var TestSplitConcatenatesTo = []struct {
		name             string
		from             string
		to               []string
		message          string
		encoder          Encoder
		shortReference   bool
		expectedSplit    []string
	}{
		{
			"message with GSM, should not split",
			from,
			[]string{to},
			message,
			NewGSM(),
			true,
			[]string{message},
		},
		{
			"message with UTF16, should split",
			from,
			[]string{to},
			message,
			NewUTF16(),
			true,
			[]string{"All of the characters that make up this message are in the GSM ", "character set."},
		},
		{
			"message with 2 byte reference bit",
			from,
			[]string{to},
			message,
			NewUTF16(),
			false,
			[]string{"All of the characters that make up this message are in the GSM ", "character set."},
		},
	}

	for _, tt := range TestSplitConcatenatesTo {
		splitter := NewSplitter()
		splitter.SetEncoder(tt.encoder)
		splitter.SetShortReference(tt.shortReference)

		SMSs, err := splitter.Split(tt.from, tt.to, tt.message)
		if err != nil {
			t.Fatalf("an error '%s' was encountered when splitting the message for test '%s'", err, tt.name)
		}

		// check SMS field correctness
		assert.Equal(t, len(tt.expectedSplit), len(SMSs))
		for idx, messagePart := range tt.expectedSplit {
			assert.Equal(t, from, SMSs[idx].from)
			assert.Equal(t, to, SMSs[idx].to)
			assert.Equal(t, messagePart, SMSs[idx].content)
		}

		// check the udh lengths
		if len(SMSs) > 1 {
			if tt.shortReference {
				assert.Equal(t, 6, len(SMSs[0].udh))
			} else {
				assert.Equal(t, 7, len(SMSs[0].udh))
			}
		}
	}
}

// this test ensures that Split fails as expected when something goes wrong
func TestSplitFails(t *testing.T) {
	const from = "from"
	const to = "to"

	var TestSplitConcatenatesTo = []struct {
		name           string
		from           string
		to             []string
		message        string
		bytes          int
		encoder        Encoder
		shortReference bool
		expectedError  error
	}{
		{
			"zero byte message, impossible split",
			from,
			[]string{to},
			"Message",
			0,
			NewUTF16(),
			false,
			ErrNotSplittable,
		},
		{
			"unencodable character",
			from,
			[]string{to},
			"Message: 你好朋友你好朋友",
			DefaultSMSBytes,
			NewGSM(),
			false,
			ErrNotEncodable,
		},
	}

	for _, tt := range TestSplitConcatenatesTo {
		splitter := NewSplitter()
		splitter.SetMessageBytes(tt.bytes)
		splitter.SetEncoder(tt.encoder)
		splitter.SetShortReference(tt.shortReference)

		SMSs, err := splitter.Split(tt.from, tt.to, tt.message)

		assert.Nil(t, SMSs)
		assert.EqualError(t, tt.expectedError, err.Error())
	}
}

// this test ensures that CheckEncodability works correctly
func TestCheckEncodability(t *testing.T) {
	var TestSplitConcatenatesTo = []struct {
		name           string
		message        string
		encoder        Encoder
		expectedResult bool
	}{
		{
			"Encodable as GSM",
			"Message",
			NewGSM(),
			true,
		},
		{
			"Encodable as UTF16",
			"Message: 你好朋友你好朋友",
			NewUTF16(),
			true,
		},
		{
			"Not Encodable as GSM",
			"Message: 你好朋友你好朋友",
			NewGSM(),
			false,
		},
	}

	for _, tt := range TestSplitConcatenatesTo {
		splitter := NewSplitter()
		splitter.SetEncoder(tt.encoder)

		encodable := splitter.CheckEncodability(tt.message)

		assert.Equal(t, tt.expectedResult, encodable)
	}
}

// this test makes sure that appendUDHs does not append a UDH to SMSs if too few are given
func TestAppendUDHsMakesNoChanges(t *testing.T) {
	const from = "from"
	const to = "to"
	const content = "content"

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
			[]SMS{newSMS(from, to, content, "")},
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
			assert.Equal(t, "", sms.udh)
		}
	}
}

// this test makes sure that appendUDHs adds a UDH correctly
func TestAppendUDHsAddsUDHWithShortReferenceNumber(t *testing.T) {
	const from = "from"
	const to = "to"
	const content = "content"

	var TestAppendUDHsAddsUDHWithShortReferenceNumber = []struct {
		name string
		SMSs []SMS
	}{
		{
			"just enough SMSs",
			[]SMS{
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
			},
		},
		{
			"loads of SMSs",
			[]SMS{
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
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

	var TestAppendUDHsAddsUDHWithLongReferenceNumber = []struct {
		name string
		SMSs []SMS
	}{
		{
			"just enough SMSs",
			[]SMS{
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
			},
		},
		{
			"loads of SMSs",
			[]SMS{
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
				newSMS(from, to, content, ""),
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
