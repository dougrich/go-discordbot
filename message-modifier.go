package discordbot

import (
	"fmt"
)

type MessageModifier interface {
	Modify() error
}

type messageModifier struct {
	message string
}

func (m *messageModifier) Modify() error {
	return nil
}

func WithMessage(format string, a ...interface{}) MessageModifier {
	return &messageModifier{
		fmt.Sprintf(format, a...),
	}
}
