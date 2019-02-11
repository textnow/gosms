package gosms

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultSMSBytes is the default SMS size in bytes
	DefaultSMSBytes             int = 140
	shortReferenceInfoElementID int = 0x00
	longReferenceInfoElementID  int = 0x08
	udhByteLengthShort          int = 6
	udhByteLengthLong           int = 7
	byteLength                  int = 8
)

// SMS structure with correctly sized message and appropriate UDH
type SMS struct {
	from    string
	to      string
	content string
	encoder string
	udh     string
}

// newSMS initializes a new SMS
func newSMS(from string, to string, content string, encoder string, udh string) SMS {
	return SMS{
		from:    from,
		to:      to,
		content: content,
		encoder: encoder,
		udh:     udh,
	}
}

// GetFrom returns the SMS's from field
func (s *SMS) GetFrom() string {
	return s.from
}

// GetTo returns the SMS's to field
func (s *SMS) GetTo() string {
	return s.to
}

// GetContent returns the SMS's content field
func (s *SMS) GetContent() string {
	return s.content
}

// GetEncoder returns the SMS's encoder field
func (s *SMS) GetEncoder() string {
	return s.encoder
}

// GetUDH returns the SMS's udh field
func (s *SMS) GetUDH() string {
	return s.udh
}

// CreateSMSPayloads generates SMSs with sizable message parts and appropriate UDHs
func CreateSMSPayloads(from string, to []string, message string, encoder Encoder, shortReference bool) []SMS {
	return CreateSMSPayloadsWithSize(from, to, message, encoder, DefaultSMSBytes, shortReference)
}

// CreateSMSPayloadsWithSize generates SMSs with sizable message parts and appropriate UDHs
func CreateSMSPayloadsWithSize(from string, to []string, message string, encoder Encoder, messageBytes int, shortReference bool) []SMS {
	var SMSs []SMS
	var messageLength int
	var messageParts []string
	var receivers string
	var udhByteLength int

	// set of symbols which compose the message
	runeSet := []rune(message)

	// append receivers
	receivers = strings.Join(to, " ")

	// short circuit for messages that don't need to be split
	singleSMS := willMessageFit(runeSet, encoder, (messageBytes*byteLength)/encoder.GetCodePointBits())
	if singleSMS {
		return []SMS{newSMS(from, receivers, message, encoder.GetEncoderName(), "")}
	}

	// determine the UDH length
	udhByteLength = udhByteLengthLong
	if shortReference {
		udhByteLength = udhByteLengthShort
	}

	// adjust message length for UDH
	messageLength = ((messageBytes - udhByteLength) * byteLength) / encoder.GetCodePointBits()

	// split message
	messageParts = SplitMessage(runeSet, encoder, messageLength)

	// create SMSs and append UDHs
	for _, messagePart := range messageParts {
		SMSs = append(SMSs, newSMS(from, receivers, messagePart, encoder.GetEncoderName(), ""))
	}
	return appendUDHs(SMSs, shortReference)
}

// appendUDHs generates UDHs for SMSs
// if messages cannot be uniquely identified, try increasing the
// size of the reference number by setting shortReference to false
func appendUDHs(SMSs []SMS, shortReference bool) []SMS {
	const timeBase = 10
	var infoElementID int
	var headerLength int
	var infoElementDataLength int
	var udhByteLength int

	// short circuit for too few SMSs
	if len(SMSs) <= 1 {
		return SMSs
	}

	// account for the length of the UDH reference number
	if shortReference {
		infoElementID = shortReferenceInfoElementID
		udhByteLength = udhByteLengthShort
	} else {
		infoElementID = longReferenceInfoElementID
		udhByteLength = udhByteLengthLong
	}

	// messageHash is not intended to be cryptographically strong,
	// but merely unique enough to identify a message.
	md5Hash := md5.New()
	microtime := time.Now().UnixNano() / int64(time.Millisecond)
	io.WriteString(md5Hash, // create an MD5 hash of
		SMSs[0].from+ // the sender
			SMSs[0].to+ // the receivers
			SMSs[0].content+ // the first message part
			strconv.FormatInt(microtime, timeBase), // the time
	) // to ensure uniqueness
	messageHash := md5Hash.Sum(nil)

	// set header and element length for UDH
	headerLength = udhByteLength - 1          // the header takes the remaining bytes of the UDH
	infoElementDataLength = udhByteLength - 3 // the header data takes the remaining bytes of the UDH

	// append UDH to messages, create SMSs
	for idx := range SMSs {
		udh := new(bytes.Buffer)
		byteOrder := binary.LittleEndian

		// construct UDH
		binary.Write(udh, byteOrder, int8(headerLength))
		binary.Write(udh, byteOrder, int8(infoElementID))
		binary.Write(udh, byteOrder, int8(infoElementDataLength))
		binary.Write(udh, byteOrder, int8(messageHash[0]))
		if !shortReference {
			binary.Write(udh, byteOrder, int8(messageHash[1]))
		}
		binary.Write(udh, byteOrder, int8(len(SMSs)))
		binary.Write(udh, byteOrder, int8(idx+1))

		SMSs[idx].udh = string(udh.Bytes())
	}

	return SMSs
}
