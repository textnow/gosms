package gosms

// SMS structure with correctly sized message and appropriate UDH
type SMS struct {
	from    string
	to      string
	content string
	udh     string
}

// newSMS initializes a new SMS
func newSMS(from string, to string, content string, udh string) SMS {
	return SMS{
		from:    from,
		to:      to,
		content: content,
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

// GetUDH returns the SMS's udh field
func (s *SMS) GetUDH() string {
	return s.udh
}
