package kumnats

type (
	// NatsMessage :nodoc:
	NatsMessage struct {
		ID     int64     `json:"id"`
		UserID int64     `json:"user_id"`
		Type   EventType `json:"type"`
		Time   string    `json:"time"`
		// deprecated field
		// keep it for now for backwards compatibility
		Body   string    `json:"body,omitempty"`

		// new fields
		Request []byte 		`json:"request"`
	}

	natsMessageWithSubject struct {
		Subject string `json:"subject"`
		Message []byte `json:"message"`
	}
)
