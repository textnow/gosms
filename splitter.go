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

// Splitter splits messages into SMS structs
type Splitter struct {
	encoder        Encoder
	messageBytes   int
	shortReference bool
}

// NewSplitter creates a new Splitter configured with default values
func NewSplitter() *Splitter {
	return &Splitter{
		encoder: nil,
		messageBytes: DefaultSMSBytes,
		shortReference: true,
	}
}

// SetEncoder sets the encoder of the Splitter
func (s *Splitter) SetEncoder(encoder Encoder) {
	s.encoder = encoder
}

// SetMessageBytes sets the messageBytes of the Splitter
func (s *Splitter) SetMessageBytes(messageBytes int) {
	s.messageBytes = messageBytes
}

// SetShortReference sets the shortReference of the Splitter
func (s *Splitter) SetShortReference(shortReference bool) {
	s.shortReference = shortReference
}

// CheckEncodability returns true if the message is encodable with the splitter's encoder and false otherwise
func (s *Splitter) CheckEncodability(message string) bool {
	return s.encoder.CheckEncodability(message)
}

// Split generates SMSs with sizable message parts and appropriate UDHs
func (s *Splitter) Split(from string, to []string, message string) ([]SMS, error) {
	var smsParts []SMS
	var messageLength int
	var messageParts []string
	var receivers string
	var udhByteLength int
	var encoder Encoder

	// use the specified encoder or auto-detect
	if s.encoder != nil {
		encoder = s.encoder
	} else {
		encoder = autoDetectEncoder(message)
	}

	// set of symbols which compose the message
	runeSet := []rune(message)

	// append receivers
	receivers = strings.Join(to, " ")

	// short circuit for messages that don't need to be split
	singleSMS, err := willMessageFit(runeSet, encoder, (s.messageBytes*byteLength)/encoder.GetCodePointBits())
	if err != nil {
		return nil, err
	}

	if singleSMS {
		return []SMS{newSMS(from, receivers, message, "")}, nil
	}

	// determine the UDH length
	udhByteLength = udhByteLengthLong
	if s.shortReference {
		udhByteLength = udhByteLengthShort
	}

	// adjust message length for UDH
	messageLength = ((s.messageBytes - udhByteLength) * byteLength) / encoder.GetCodePointBits()

	// split message
	messageParts, err = SplitMessage(runeSet, encoder, messageLength)
	if err != nil {
		return nil, err
	}

	// create SMS parts and append UDHs
	for _, messagePart := range messageParts {
		smsParts = append(smsParts, newSMS(from, receivers, messagePart, ""))
	}
	return appendUDHs(smsParts, s.shortReference), nil
}

// appendUDHs generates UDHs for SMS parts
// if messages cannot be uniquely identified, try increasing the
// size of the reference number by setting shortReference to false
func appendUDHs(smsParts []SMS, shortReference bool) []SMS {
	const timeBase = 10
	var infoElementID int
	var headerLength int
	var infoElementDataLength int
	var udhByteLength int

	// short circuit for too few SMS parts
	if len(smsParts) <= 1 {
		return smsParts
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
		smsParts[0].from+ // the sender
			smsParts[0].to+ // the receivers
			smsParts[0].content+ // the first message part
			strconv.FormatInt(microtime, timeBase), // the time
	) // to ensure uniqueness
	messageHash := md5Hash.Sum(nil)

	// set header and element length for UDH
	headerLength = udhByteLength - 1          // the header takes the remaining bytes of the UDH
	infoElementDataLength = udhByteLength - 3 // the header data takes the remaining bytes of the UDH

	// append UDH to messages, create SMS parts
	for idx := range smsParts {
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
		binary.Write(udh, byteOrder, int8(len(smsParts)))
		binary.Write(udh, byteOrder, int8(idx+1))

		smsParts[idx].udh = udh.String()
	}

	return smsParts
}

func autoDetectEncoder(message string) Encoder {
	runeSet := []rune(message)
	for _, char := range runeSet {
		_, isGSM := gsmCodePoints[char]
		if !isGSM {
			return NewUTF16()
		}
	}
	return NewGSM()
}

