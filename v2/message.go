package kumnats

import (
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
		Body   string    `json:"body,omitempty"`
		Time   string    `json:"time"`
	}

	natsMessageWithSubject struct {
		Subject string `json:"subject"`
		Message []byte `json:"message"`
	}

	NatsMessageWithOldData struct {
		NatsMessage
		OldData string `json:"old_data,omitempty"`
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

func (n *NatsMessageWithOldData) ParseFromBytes(data []byte) (err error) {
	err = tapao.Unmarshal(data, &n, tapao.FallbackWith(tapao.JSON))
	if err != nil {
		logrus.WithField("data", string(data)).Error(err)
	}
	return
}
