package add

import (
	"context"

	"github.com/dougrich/go-discordbot"
)

type CommandPackage struct{}

func (CommandPackage) Name() string {
	return "Add"
}

func (CommandPackage) Register(bot *discordbot.Bot) error {
	bot.AddCommand(
		"add",
		[]discordbot.Option{
			{
				Type: discordbot.OptionNumber,
				Name: "a",
			},
			{
				Type: discordbot.OptionNumber,
				Name: "b",
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
	return nil
}
