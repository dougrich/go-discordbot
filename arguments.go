package discordbot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Wrapper for a slash commands arguments
type Arguments struct {
	data discordgo.ApplicationCommandInteractionData
}

// Scans the slash command arguments into fields
func (a *Arguments) Scan(dest ...interface{}) error {
	opts := a.data.Options
	if len(opts) == 0 {
		return nil
	}
	if opts[0].Type == discordgo.ApplicationCommandOptionSubCommand {
		opts = opts[0].Options
	}
	for i, d := range dest {
		if i >= len(opts) {
			return nil
		}
		switch ptr := d.(type) {
		case *int64:
			*ptr = opts[i].IntValue()
		default:
			return fmt.Errorf("discordbot: unfamiliar scan type in arguments %T", ptr)
		}
	}
	return nil
}
