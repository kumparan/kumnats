package kumnats

type (
	// NatsMessage :nodoc:
	NatsMessage struct {
		ID     int64     `json:"id"`
		UserID int64     `json:"user_id"`
		Type   EventType `json:"type"`
		Body   string    `json:"body,omitempty"`
		Time   string    `json:"time"`
	}

	natsMessageWithSubject struct {
		Subject string      `json:"subject"`
		Message interface{} `json:"message"`
	}
)
