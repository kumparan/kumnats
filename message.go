package kumnats

type (
	// NatsMessage :nodoc:
	NatsMessage struct {
		ID     int64     `json:"id" msgpack:"id"`
		UserID int64     `json:"user_id" msgpack:"user_id"`
		Type   EventType `json:"type" msgpack:"type"`
		Body   string    `json:"body,omitempty" msgpack:"body,omitempty"`
		Time   string    `json:"time" msgpack:"time"`
	}

	natsMessageWithSubject struct {
		Subject string `json:"subject" msgpack:"subject"`
		Message []byte `json:"message" msgpack:"message"`
	}
)
