package kumnats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kumparan/tapao"
)

func TestNatsMessage_ParseBytes(t *testing.T) {
	msg := &NatsMessage{
		ID:     1573808900293581737,
		UserID: 1573808900293581738,
		Type:   "anu",
		Body:   "ea",
		Time:   time.Now().Format(time.RFC3339Nano),
	}
	data, err := tapao.Marshal(msg)
	assert.NoError(t, err)

	m := new(NatsMessage)
	err = m.ParseFromBytes(data)
	assert.NoError(t, err)
	assert.Equal(t, msg.ID, m.ID)
	assert.Equal(t, msg.UserID, m.UserID)
	assert.Equal(t, msg.Type, m.Type)
	assert.Equal(t, msg.Body, m.Body)
}

func TestNatsMessageWithOldData_ParseBytes(t *testing.T) {
	natsMsg := NatsMessage{
		ID:     1573808900293581737,
		UserID: 1573808900293581738,
		Type:   "anu",
		Body:   "ea",
		Time:   time.Now().Format(time.RFC3339Nano),
	}

	natsMsgWithOldData := NatsMessageWithOldData{
		NatsMessage: natsMsg,
		OldData:     "zzz",
	}

	data, err := tapao.Marshal(natsMsgWithOldData)
	assert.NoError(t, err)

	// m := new(NatsMessage)
	km := new(NatsMessageWithOldData)

	err = km.ParseFromBytes(data)
	assert.NoError(t, err)

	assert.Equal(t, natsMsg.ID, km.ID)
	assert.Equal(t, natsMsg.UserID, km.UserID)
	assert.Equal(t, natsMsg.Type, km.Type)
	assert.Equal(t, natsMsg.Body, km.Body)
	assert.Equal(t, natsMsgWithOldData.OldData, km.OldData)
}

func TestNatsMessage_ParseFromJSON(t *testing.T) {
	msg := &NatsMessage{
		ID:     1573808900293581737,
		UserID: 1573808900293581738,
		Type:   "anu",
		Body:   "ea",
		Time:   time.Now().Format(time.RFC3339Nano),
	}
	data, err := tapao.Marshal(msg, tapao.FallbackWith(tapao.JSON))
	assert.NoError(t, err)

	m := new(NatsMessage)
	err = m.ParseFromJSON(string(data))
	assert.NoError(t, err)
	assert.Equal(t, msg.ID, m.ID)
	assert.Equal(t, msg.UserID, m.UserID)
	assert.Equal(t, msg.Type, m.Type)
	assert.Equal(t, msg.Body, m.Body)
}

func TestNatsMessage_ToJSON(t *testing.T) {
	msg := &NatsMessage{
		ID:     1573808900293581737,
		UserID: 1573808900293581738,
		Type:   "anu",
		Body:   "ea",
		Time:   time.Now().Format(time.RFC3339Nano),
	}

	data, err := msg.ToJSON()
	assert.NoError(t, err)
	assert.NotNil(t, data)
}
