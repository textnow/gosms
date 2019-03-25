package gosms

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// this test ensures that willMessageFit functions as expected. willMessageFit should
// return false for messages that are longer than messageLength and true otherwise.
func TestWillMessageFit(t *testing.T) {
	var TestWillMessageFit = []struct {
		name          string
		message       []rune
		encoder       Encoder
		messageLength int
		expected      bool
	}{
		{
			"Not too long, no special GSM characters",
			[]rune("Not too long!"),
			NewGSM(),
			13,
			true,
		},
		{
			"Not too long, no special unicode characters",
			[]rune("Not too long! 你好朋友"),
			NewUTF16(),
			18,
			true,
		},
		{
			"Not too long, with special GSM character",
			[]rune("Not too long ~"),
			NewGSM(),
			15,
			true,
		},
		{
			"Not too long, with special unicode character",
			[]rune("Not too long 🙂"),
			NewUTF16(),
			15,
			true,
		},
		{
			"Too long, no special GSM characters",
			[]rune("Oops! Too long"),
			NewGSM(),
			13,
			false,
		},
		{
			"Too long, no special unicode characters",
			[]rune("Oops! Too long 你好朋友"),
			NewUTF16(),
			18,
			false,
		},
		{
			"Too long, with special GSM character",
			[]rune("Oops! Too long ~"),
			NewGSM(),
			16,
			false,
		},
		{
			"Too long, with special unicode character",
			[]rune("Oops! Too long 🙂"),
			NewUTF16(),
			16,
			false,
		},
	}

	for _, tt := range TestWillMessageFit {
		fit, err := willMessageFit(tt.message, tt.encoder, tt.messageLength)
		if err != nil {
			t.Fatalf("an error '%s' was encountered when checking message fit for test '%s'", err, tt.name)
		}
		assert.Equal(t, tt.expected, fit)
	}
}

func TestWillMessageFitFailsForUnencodableMessage(t *testing.T) {
	message := []rune("Message: 你好朋友你好朋友")
	encoder := NewGSM()

	fit, err := willMessageFit(message, encoder, DefaultSMSBytes)

	assert.False(t, fit)
	assert.EqualError(t, ErrNotEncodable, err.Error())
}

// this test ensures that SplitMessage functions as expected. Messages should be split into
// chunks of `messageLength` code points. Words should not be split. Failure returns nil.
func TestSplitMessage(t *testing.T) {
	var TestSplitMessage = []struct {
		name          string
		message       []rune
		encoder       Encoder
		messageLength int
		expected      []string
	}{
		{
			"Fully pack string without splitting",
			[]rune("Don't split me!"),
			NewGSM(),
			15,
			[]string{
				"Don't split me!",
			},
		},
		{
			"split before and after spaces",
			[]rune("Split after space. Split before space. Good!"),
			NewGSM(),
			19,
			[]string{
				"Split after space. ",
				"Split before space.",
				" Good!",
			},
		},
		{
			"split after, but not before, punctuation",
			[]rune("Split after punctuation, dont before punctuation. Good!"),
			NewGSM(),
			24,
			[]string{
				"Split after punctuation,",
				" dont before ",
				"punctuation. Good!",
			},
		},
		{
			"Don't split in the middle of a word with available split points",
			[]rune("Are you an antidisestablishmentarian?"),
			NewGSM(),
			26,
			[]string{
				"Are you an ",
				"antidisestablishmentarian?",
			},
		},
		{
			"Split long strings if there are not available split points",
			[]rune("Antidisestablishmentarianism"),
			NewGSM(),
			10,
			[]string{
				"Antidisest",
				"ablishment",
				"arianism",
			},
		},
		{
			"Count code points, not characters or bytes (GSM)",
			[]rune("[18 code points]New line"),
			NewGSM(),
			18,
			[]string{
				"[18 code points]",
				"New line",
			},
		},
		{
			"Count code points, not characters or bytes (Unicode)",
			[]rune("😅18 code points😅 New line"),
			NewUTF16(),
			18,
			[]string{
				"😅18 code points😅",
				" New line",
			},
		},
		{
			"Don't split special GSM characters",
			[]rune("this string is 37 code points long ~"),
			NewGSM(),
			36,
			[]string{
				"this string is 37 code points long ",
				"~",
			},
		},
		{
			"Don't split special unicode characters",
			[]rune("this string is 37 code points long 😁"),
			NewUTF16(),
			36,
			[]string{
				"this string is 37 code points long ",
				"😁",
			},
		},
		{
			"Correctly split an empty message",
			[]rune(""),
			NewGSM(),
			0,
			[]string{""},
		},
		{
			"Don't count special GSM characters as special in unicode",
			[]rune("🙂[]🙂"),
			NewUTF16(),
			2,
			[]string{
				"🙂",
				"[]",
				"🙂",
			},
		},
		{
			"Fail on generally impossible split",
			[]rune("X"),
			NewGSM(),
			0,
			nil,
		},
		{
			"Fail on impossible GSM split",
			[]rune("~"),
			NewGSM(),
			1,
			nil,
		},
		{
			"Fail on impossible unicode split",
			[]rune("🙂"),
			NewUTF16(),
			1,
			nil,
		},
	}

	for _, tt := range TestSplitMessage {
		messages, err := SplitMessage(tt.message, tt.encoder, tt.messageLength)

		// did SplitMessage fail?
		if err != nil {
			assert.Equal(t, tt.expected, messages)
			continue
		}

		// check expected vs actual
		assert.Equal(t, len(tt.expected), len(messages))
		for idx := range messages {
			assert.Equal(t, tt.expected[idx], messages[idx])
		}
		assert.Equal(t, string(tt.message), strings.Join(messages, ""))
	}
}

func TestSplitMessageFailsForUnencodableMessage(t *testing.T) {
	message := []rune("Message: 你好朋友你好朋友")
	encoder := NewGSM()

	split, err := SplitMessage(message, encoder, DefaultSMSBytes)

	assert.Nil(t, split)
	assert.EqualError(t, ErrNotEncodable, err.Error())
}
