package utils

import "errors"

var (
	ErrNotEnoughGuesses = errors.New("not enough guesses available")
	ErrRoomClosed       = errors.New("room is already closed")
	ErrLimitExceeded    = errors.New("limit exceeded")
)
