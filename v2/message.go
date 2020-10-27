package kumnats

import (
	"time"

	"github.com/kumparan/tapao"
	"github.com/sirupsen/logrus"
)

type (
	// MessagePayload :nodoc:
	MessagePayload interface {
		ParseFromBytes(data []byte) error
	}

	// NatsMessage :nodoc:
	NatsMessage struct {
		ID     int64     `json:"id"`
		UserID int64     `json:"user_id"`
		Type   EventType `json:"type"`
		Time   string    `json:"time"`
		// deprecated field
		// keep it for now for backwards compatibility
		Body string `json:"body,omitempty"`

		// new fields
		Request []byte `json:"request"`
	}

	natsMessageWithSubject struct {
		Subject string `json:"subject"`
		Message []byte `json:"message"`
	}

	// NatsMessageWithOldData :nodoc:
	NatsMessageWithOldData struct {
		NatsMessage
		OldData string `json:"old_data,omitempty"`
	}

	// AuditLogMessage :nodoc:
	AuditLogMessage struct {
		ServiceName    string    `json:"service_name"`
		UserID         int64     `json:"user_id"`
		AuditableType  string    `json:"auditable_type"`
		AuditableID    string    `json:"auditable_id"`
		Action         string    `json:"action"`
		AuditedChanges string    `json:"audited_changes"`
		OldData        string    `json:"old_data,omitempty"`
		NewData        string    `json:"new_data,omitempty"`
		CreatedAt      time.Time `json:"created_at"`
	}
)

// ParseFromBytes implementation of NatsMessage
func (m *NatsMessage) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &m, tapao.FallbackWith(tapao.JSON))
	if err != nil {
		logrus.WithField("data", string(data)).Error(err)
	}
	return
}

// ParseFromBytes implementation of NatsMessageWithOldData
func (n *NatsMessageWithOldData) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &n, tapao.FallbackWith(tapao.JSON))
	if err != nil {
		logrus.WithField("data", string(data)).Error(err)
	}
	return
}

// ParseFromBytes implementation of AuditLogMessage
func (m *AuditLogMessage) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &m, tapao.FallbackWith(tapao.JSON))
	if err != nil {
		logrus.WithField("data", string(data)).Error(err)
	}
	return
}
