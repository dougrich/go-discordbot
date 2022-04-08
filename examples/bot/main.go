package main

import (
	"os"

	"github.com/dougrich/go-discordbot"
	"github.com/dougrich/go-discordbot/examples/cmd/add"
)

func main() {
	discordbot.New(
		discordbot.BotOptions{
			Token:          os.Getenv("DEVELOPMENT_BOT_TOKEN"),
			GuildID:        os.Getenv("DEVELOPMENT_GUILD_ID"),
			RemoveCommands: true,
		},
	).Include(&add.CommandPackage{}).Start()
}
