package kumnats

import (
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/kumparan/go-lib/utils"
	"github.com/kumparan/tapao"
	"github.com/pkg/errors"
)

type (
	// NatsEvent :nodoc:
	NatsEvent struct {
		ID     int64
		UserID int64
		Type   EventType
	}

	// NatsEventMessage :nodoc:
	NatsEventMessage struct {
		NatsEvent *NatsEvent
		Body      string
		OldBody   string
		Request   []byte
		Error     error
	}
)

// GetID :nodoc:
func (n *NatsEvent) GetID() int64 {
	if n == nil {
		return 0
	}
	return n.ID
}

// GetUserID :nodoc:
func (n *NatsEvent) GetUserID() int64 {
	if n == nil {
		return 0
	}
	return n.UserID
}

// GetType :nodoc:
func (n *NatsEvent) GetType() EventType {
	if n == nil {
		return ""
	}
	return n.Type
}

// NewNatsEventMessage :nodoc:
func NewNatsEventMessage() *NatsEventMessage {
	return &NatsEventMessage{}
}

// Build :nodoc:
func (n *NatsEventMessage) Build() (data []byte, err error) {
	if n.Error != nil {
		return nil, n.Error
	}

	if n.NatsEvent == nil {
		n.wrapError(errors.New("empty nats nats event"))
		return nil, n.Error
	}

	var message []byte
	switch {
	case n.OldBody != "":
		message, err = tapao.Marshal(NatsMessageWithOldData{
			NatsMessage: NatsMessage{
				ID:      n.NatsEvent.ID,
				UserID:  n.NatsEvent.UserID,
				Type:    n.NatsEvent.Type,
				Time:    time.Now().Format(time.RFC3339Nano),
				Body:    n.Body,
				Request: n.Request,
			},
			OldData: n.OldBody,
		})
	default:
		message, err = tapao.Marshal(NatsMessage{
			ID:      n.NatsEvent.ID,
			UserID:  n.NatsEvent.UserID,
			Type:    n.NatsEvent.Type,
			Time:    time.Now().Format(time.RFC3339Nano),
			Body:    n.Body,
			Request: n.Request,
		})
	}
	if err != nil {
		n.wrapError(err)
		return nil, n.Error
	}

	return message, nil
}

// WithEvent :nodoc:
func (n *NatsEventMessage) WithEvent(e *NatsEvent) *NatsEventMessage {
	if e.GetID() <= 0 {
		n.wrapError(errors.New("empty id"))
		return n
	}

	if e.GetUserID() == 0 {
		n.wrapError(errors.New("empty user id"))
		return n
	}
	if e.GetType() == "" {
		n.wrapError(errors.New("empty NatsEvent type"))
		return n
	}

	n.NatsEvent = e
	return n
}

// WithBody :nodoc:
func (n *NatsEventMessage) WithBody(body interface{}) *NatsEventMessage {
	n.Body = utils.Dump(body)
	return n
}

// WithOldBody :nodoc:
func (n *NatsEventMessage) WithOldBody(body interface{}) *NatsEventMessage {
	n.OldBody = utils.Dump(body)
	return n
}

// WithRequest :nodoc:
func (n *NatsEventMessage) WithRequest(req proto.Message) *NatsEventMessage {
	b, err := tapao.Marshal(req, tapao.With(tapao.Protobuf))
	if err != nil {
		n.wrapError(err)
		return n
	}

	n.Request = b
	return n
}

func (n *NatsEventMessage) wrapError(err error) {
	if n.Error != nil {
		n.Error = errors.Wrap(n.Error, err.Error())
		return
	}
	n.Error = err
}
