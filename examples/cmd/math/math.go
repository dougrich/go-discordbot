package math

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/dougrich/go-discordbot"
)

type CommandPackage struct{}

func (CommandPackage) Name() string {
	return "Add"
}

func (CommandPackage) Register(bot *discordbot.Bot) error {
	bot.AddCommand(
		"math",
		[]*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "adds a + b",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "a",
						Description: "integer a",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "b",
						Description: "integer b",
						Required:    true,
					},
				},
			},
		},
		"adds a + b",
		func(ctx context.Context, args *discordbot.Arguments) error {
			var a, b int64
			if err := args.Scan(&a, &b); err != nil {
				return err
			}
			return bot.Respond(ctx, discordbot.WithMessage("%d+%d=%d", a, b, a+b))
		},
	)
	bot.AddCommand(
		"math",
		[]*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "sub",
				Description: "subtract a - b",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "a",
						Description: "integer a",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Name:        "b",
						Description: "integer b",
						Required:    true,
					},
				},
			},
		},
		"subtract a - b",
		func(ctx context.Context, args *discordbot.Arguments) error {
			var a, b int64
			if err := args.Scan(&a, &b); err != nil {
				return err
			}
			return bot.Respond(ctx, discordbot.WithMessage("%d-%d=%d", a, b, a-b))
		},
	)
	return nil
}
