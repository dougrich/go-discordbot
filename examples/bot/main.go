package main

import (
	"os"

	"github.com/dougrich/go-discordbot"
	"github.com/dougrich/go-discordbot/examples/cmd/add"
	"github.com/dougrich/go-discordbot/examples/cmd/math"
)

func main() {
	b := discordbot.New(
		discordbot.BotOptions{
			Token:          os.Getenv("DEVELOPMENT_BOT_TOKEN"),
			GuildID:        os.Getenv("DEVELOPMENT_GUILD_ID"),
			RemoveCommands: true,
		},
	)
	b.Include(&add.CommandPackage{})
	b.Include(&math.CommandPackage{})
	b.Start()
}
