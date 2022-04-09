package discordbot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type MessageModifier interface {
	Modify(*discordgo.InteractionResponse) error
}

type messageModifier struct {
	message string
}

func (m *messageModifier) Modify(i *discordgo.InteractionResponse) error {
	i.Data.Content = m.message
	return nil
}

func WithMessage(format string, a ...interface{}) MessageModifier {
	return &messageModifier{
		fmt.Sprintf(format, a...),
	}
}
