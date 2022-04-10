package discordbot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type MessageModifier interface {
	ModifyInteraction(*Bot, *discordgo.InteractionResponse) error
	ModifyMessage(*Bot, *discordgo.MessageSend) error
	ModifyMessageAfter(*Bot, *discordgo.Message) error
}

// message modifier

type messageModifier struct {
	message string
}

func (m *messageModifier) ModifyInteraction(_ *Bot, i *discordgo.InteractionResponse) error {
	i.Data.Content = m.message
	return nil
}

func (m *messageModifier) ModifyMessage(_ *Bot, msg *discordgo.MessageSend) error {
	msg.Content = m.message
	return nil
}

func (m *messageModifier) ModifyMessageAfter(*Bot, *discordgo.Message) error {
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

func (m *embedModifier) ModifyInteraction(_ *Bot, i *discordgo.InteractionResponse) error {
	i.Data.Embeds = append(i.Data.Embeds, m.embed)
	return nil
}

func (m *embedModifier) ModifyMessage(_ *Bot, msg *discordgo.MessageSend) error {
	if msg.Embed != nil {
		return errors.New("Embed already exists")
	}
	msg.Embed = m.embed
	return nil
}

func (m *embedModifier) ModifyMessageAfter(_ *Bot, msg *discordgo.Message) error {
	return nil
}

func WithEmbed(embed *discordgo.MessageEmbed) MessageModifier {
	return &embedModifier{
		embed,
	}
}

// reaction modifier

type reactionModifier struct {
	reaction string
}

func (m *reactionModifier) ModifyInteraction(*Bot, *discordgo.InteractionResponse) error {
	return nil
}

func (m *reactionModifier) ModifyMessage(*Bot, *discordgo.MessageSend) error {
	return nil
}

func (m *reactionModifier) ModifyMessageAfter(bot *Bot, msg *discordgo.Message) error {
	return bot.session.MessageReactionAdd(msg.ChannelID, msg.ID, m.reaction)
}

func WithReaction(reaction string) MessageModifier {
	return &reactionModifier{
		reaction,
	}
}
