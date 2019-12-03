package kumnats

import "errors"

var (
	// ErrBadUnmarshalResult given when unmarshal result from a message's Data is not as intended
	ErrBadUnmarshalResult = errors.New("kumnatserr: bad unmarshal result")
	// ErrGiveUpProcessingMessagePayload given when message's payload(data) is already processed x times, but always failed
	ErrGiveUpProcessingMessagePayload = errors.New("kumnatserr: give up processing message payload")
	// ErrNilMessagePayload given when message's payload(data) is nil
	ErrNilMessagePayload = errors.New("kumnatserr: nil message payload given")
)
