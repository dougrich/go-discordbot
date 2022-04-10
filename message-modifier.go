package discordbot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type MessageModifier interface {
	Modify(*discordgo.InteractionResponse) error
}

// message modifier

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

// embed modifier

type embedModifier struct {
	embed *discordgo.MessageEmbed
}

func (m *embedModifier) Modify(i *discordgo.InteractionResponse) error {
	i.Data.Embeds = append(i.Data.Embeds, m.embed)
	return nil
}

func WithEmbed(embed *discordgo.MessageEmbed) MessageModifier {
	return &embedModifier{
		embed,
	}
}
