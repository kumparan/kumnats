package kumnats

import (
	"testing"
	"time"

	"github.com/kumparan/go-lib/utils"
	"github.com/kumparan/tapao"
	"github.com/kumparan/tapao/pb"
	"github.com/stretchr/testify/assert"
)

func TestNatsEventMessage_WithEvent(t *testing.T) {
	tests := []struct {
		Name          string
		Given         *NatsEvent
		ExpectedError bool
	}{
		{
			Name: "success",
			Given: &NatsEvent{
				ID:     123,
				UserID: 123,
				Type:   "type",
			},
			ExpectedError: false,
		},
		{
			Name: "empty id",
			Given: &NatsEvent{
				UserID: 123,
				Type:   "test",
			},
			ExpectedError: true,
		},
		{
			Name: "empty user",
			Given: &NatsEvent{
				ID:   123,
				Type: "test",
			},
			ExpectedError: true,
		},
		{
			Name: "empty type",
			Given: &NatsEvent{
				ID:     123,
				UserID: 123,
			},
			ExpectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result := NewNatsEventMessage().WithEvent(test.Given)
			if test.ExpectedError {
				assert.Error(t, result.Error)
				assert.Nil(t, result.NatsEvent)
				return
			}
			assert.EqualValues(t, test.Given, result.NatsEvent)
		})
	}
}

func TestNatsEventMessage_WithBody(t *testing.T) {
	body := []string{"test"}
	result := NewNatsEventMessage().WithBody(body)
	assert.NoError(t, result.Error)
	assert.Equal(t, utils.Dump(body), result.Body)
}

func TestNatsEventMessage_WithOldBody(t *testing.T) {
	body := []string{"test"}
	result := NewNatsEventMessage().WithOldBody(body)
	assert.NoError(t, result.Error)
	assert.Equal(t, utils.Dump(body), result.OldBody)
}

func TestNatsEventMessage_WithRequest(t *testing.T) {
	body := &pb.FindByIDRequest{Id: 123}
	result := NewNatsEventMessage().WithRequest(body)
	assert.NoError(t, result.Error)

	var requestResult pb.FindByIDRequest
	err := tapao.Unmarshal(result.Request, &requestResult, tapao.With(tapao.Protobuf))
	assert.NoError(t, err)
	assert.Equal(t, body.Id, requestResult.GetId())
}

func TestNatsEventMessage_Build(t *testing.T) {
	event := &NatsEvent{
		ID:     1,
		UserID: 1,
		Type:   "type",
	}
	body := []string{"test"}
	oldBody := []string{"old test"}
	req := &pb.Greeting{
		Id:        123,
		Name:      "hey",
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}

	t.Run("success", func(t *testing.T) {
		message, err := NewNatsEventMessage().
			WithEvent(event).
			WithBody(body).
			WithRequest(req).
			Build()
		assert.NoError(t, err)
		assert.NotNil(t, message)

		var result NatsMessage
		err = tapao.Unmarshal(message, &result, tapao.FallbackWith(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, event.ID, result.ID)
		assert.Equal(t, event.UserID, result.UserID)
		assert.Equal(t, event.Type, result.Type)
		assert.Equal(t, utils.Dump(body), result.Body)

		var requestResult pb.Greeting
		err = tapao.Unmarshal(result.Request, &requestResult, tapao.With(tapao.Protobuf))
		assert.NoError(t, err)
		assert.Equal(t, req.Id, requestResult.Id)
		assert.Equal(t, req.Name, requestResult.Name)
		assert.Equal(t, req.CreatedAt, requestResult.CreatedAt)
		assert.Equal(t, req.UpdatedAt, requestResult.UpdatedAt)
	})

	t.Run("success with old Body", func(t *testing.T) {
		message, err := NewNatsEventMessage().
			WithEvent(event).
			WithBody(body).
			WithOldBody(oldBody).
			WithRequest(req).
			Build()
		assert.NoError(t, err)
		assert.NotNil(t, message)

		var result NatsMessageWithOldData
		err = tapao.Unmarshal(message, &result, tapao.FallbackWith(tapao.JSON))
		assert.NoError(t, err)
		assert.Equal(t, event.ID, result.ID)
		assert.Equal(t, event.UserID, result.UserID)
		assert.Equal(t, event.Type, result.Type)
		assert.Equal(t, utils.Dump(body), result.Body)
		assert.Equal(t, utils.Dump(oldBody), result.OldData)

		var requestResult pb.Greeting
		err = tapao.Unmarshal(result.Request, &requestResult, tapao.With(tapao.Protobuf))
		assert.NoError(t, err)
		assert.Equal(t, req.Id, requestResult.Id)
		assert.Equal(t, req.Name, requestResult.Name)
		assert.Equal(t, req.CreatedAt, requestResult.CreatedAt)
		assert.Equal(t, req.UpdatedAt, requestResult.UpdatedAt)
	})

	t.Run("missing nats event", func(t *testing.T) {
		message, err := NewNatsEventMessage().Build()
		assert.Error(t, err)
		assert.Nil(t, message)

		events := []*NatsEvent{
			nil,
			{
				UserID: 123,
				Type:   "test",
			},
			{
				ID:   123,
				Type: "test",
			},
			{
				ID:     1234,
				UserID: 123,
			},
		}
		for _, e := range events {
			message, err = NewNatsEventMessage().WithEvent(e).Build()
			assert.Error(t, err)
			assert.Nil(t, message)
		}
	})
}
