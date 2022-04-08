package discordbot

import (
	"context"
)

type Bot struct {
}

type BotOptions struct {
	Token          string
	GuildID        string
	RemoveCommands bool
}

type Option struct {
	Type OptionType
	Name string
}

type OptionType uint8

const (
	OptionSubCommand      OptionType = 1
	OptionSubCommandGroup OptionType = 2
	OptionString          OptionType = 3
	OptionInteger         OptionType = 4
	OptionBoolean         OptionType = 5
	OptionUser            OptionType = 6
	OptionChannel         OptionType = 7
	OptionRole            OptionType = 8
	OptionMentionable     OptionType = 9
	OptionNumber          OptionType = 10
	OptionAttachment      OptionType = 11
)

type CommandHandler func(ctx context.Context, args *Arguments) error

type BotPackage interface {
	Name() string
	Register(b *Bot) error
}

func New(options BotOptions) *Bot {
	return &Bot{}
}

func (b *Bot) AddCommand(
	commandname string,
	options []Option,
	description string,
	handler CommandHandler,
) {

}

func (b *Bot) Respond(ctx context.Context, mods ...MessageModifier) error {
	return nil
}

func (b *Bot) Include(BotPackage) *Bot {
	return b
}

func (b *Bot) Start() {

}
