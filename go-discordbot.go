package discordbot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	// context key for interaction
	ctxKeyInteraction = "interaction"
)

// instance of a discordbot
type Bot struct {
	// any error that occured in registration
	ErrRegistration error

	// the options passed in
	options BotOptions
	// the commands to initialize
	commands []*discordgo.ApplicationCommand
	// the matchers to use for help and determining which handler to use
	matchers []commandMatcher
	// the session of the bot
	session *discordgo.Session
	// any deferred logic for after the bot cleans up
	defers []func()
}

type BotOptions struct {
	Token          string
	GuildID        string
	RemoveCommands bool
}

type CommandHandler func(ctx context.Context, args *Arguments) error
type commandMatcher struct {
	commandname string
	options     []*discordgo.ApplicationCommandOption
	description string
	cmd         *discordgo.ApplicationCommand
	handler     CommandHandler
}

func (c commandMatcher) matches(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) bool {
	d := i.ApplicationCommandData()
	isSubcommand := len(c.cmd.Options) > 0 && c.cmd.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand
	isInteractionSubcommand := len(d.Options) > 0 && d.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand
	// shortcircuit
	if isSubcommand != isInteractionSubcommand || c.cmd.Name != d.Name {
		return false
	}

	// either we're not a subcommand, or the subcommand name matches
	return !isSubcommand || c.cmd.Options[0].Name == d.Options[0].Name
}

type BotPackage interface {
	Name() string
	Register(b *Bot) error
}

func New(options BotOptions) *Bot {
	b := &Bot{
		options: options,
	}
	b.AddCommand(
		"help",
		nil,
		"provides documentation on how the slash commands work",
		b.helpInteraction,
	)
	return b
}

func (bot *Bot) Defer(later func()) {
	bot.defers = append(bot.defers, later)
}

func (bot *Bot) AddCommand(
	commandname string,
	options []*discordgo.ApplicationCommandOption,
	description string,
	handler CommandHandler,
) {
	if bot.ErrRegistration != nil {
		return
	}
	cmd := createCommand(commandname, options, description)

	bot.matchers = append(bot.matchers, commandMatcher{
		commandname,
		options,
		description,
		cmd,
		handler,
	})
	isSubcommand := len(cmd.Options) > 0 && cmd.Options[0].Type == discordgo.ApplicationCommandOptionSubCommand
	for i, c := range bot.commands {
		// special case for subcommands
		if c.Name == cmd.Name && isSubcommand {
			for _, o := range c.Options {
				if o.Type == discordgo.ApplicationCommandOptionSubCommand && o.Name == cmd.Options[0].Name {
					bot.ErrRegistration = fmt.Errorf("Failed to register \"/%s %s\", subcommand is already registered", c.Name, cmd.Options[0].Name)
					return
				}
			}
			c.Options = append(c.Options, cmd.Options[0])
			bot.commands[i] = c
			return
		} else if c.Name == cmd.Name {
			bot.ErrRegistration = fmt.Errorf("Failed to register \"/%s\", command is already registered", c.Name)
			return
		}
	}
	bot.commands = append(bot.commands, cmd)
}

func (bot *Bot) Respond(ctx context.Context, mods ...MessageModifier) error {
	i := ctx.Value(ctxKeyInteraction).(*discordgo.InteractionCreate)
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}
	for _, m := range mods {
		if err := m.Modify(response); err != nil {
			return err
		}
	}
	return bot.session.InteractionRespond(i.Interaction, response)
}

func (b *Bot) Include(p BotPackage) *Bot {
	if b.ErrRegistration == nil {
		err := p.Register(b)
		if err != nil {
			b.ErrRegistration = &errRegistration{p.Name(), err}
		} else if b.ErrRegistration != nil {
			b.ErrRegistration = &errRegistration{p.Name(), err}
		}
	}
	return b
}

func (bot *Bot) Start() {
	if bot.ErrRegistration != nil {
		log.Panicf("Invalid registration:\n%v", bot.ErrRegistration)
	}

	defer bot.runDefers()

	log.Printf("Starting bot %s", bot.options.Token)
	if bot.options.GuildID != "" {
		log.Printf(">> scoped to %s", bot.options.GuildID)
	}

	s, err := discordgo.New("Bot " + bot.options.Token)
	if err != nil {
		log.Panicf("Invalid bot parameters: %v", err)
	}
	bot.session = s

	err = s.Open()
	if err != nil {
		log.Panicf("Unable to start discord session: %v", err)
	}
	defer s.Close()
	registeredCommands := make([]*discordgo.ApplicationCommand, len(bot.commands))
	for i, v := range bot.commands {
		log.Printf("Trying to create %v as %s", v, s.State.User.ID)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, bot.options.GuildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// register callbacks
	s.AddHandler(bot.handleInteraction)

	// wait until we're done
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("starting shutdown from signal")

	// optionally remove commands
	if bot.options.RemoveCommands {
		log.Println("Removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, bot.options.GuildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}

func (bot *Bot) runDefers() {
	for _, d := range bot.defers {
		d()
	}
}

func (bot *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKeyInteraction, i)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	arguments := &Arguments{
		data: i.ApplicationCommandData(),
	}
	for _, matcher := range bot.matchers {
		if !matcher.matches(s, i) {
			continue
		}
		matcher.handler(ctx, arguments)
		return
	}
	// unmatched
	log.Print("Unmatched command")
}

func (bot *Bot) helpInteraction(ctx context.Context, _ *Arguments) error {
	var sb strings.Builder
	for _, m := range bot.matchers {
		fmt.Fprintf(&sb, "\n**/%s", m.commandname)
		for _, o := range m.options {
			switch o.Type {
			case discordgo.ApplicationCommandOptionInteger:
				fmt.Fprintf(&sb, "\t%s:integer", o.Name)
			default:
				fmt.Fprintf(&sb, "\t%s", o.Name)
			}
		}
		fmt.Fprintf(&sb, "**\n\t%s\n", m.description)
		for _, o := range m.options {
			fmt.Fprintf(&sb, "\t*")
			switch o.Type {
			case discordgo.ApplicationCommandOptionInteger:
				fmt.Fprintf(&sb, "%s:integer", o.Name)
			default:
				fmt.Fprintf(&sb, "%s", o.Name)
			}
			fmt.Fprintf(&sb, "*\t\t%s\n", o.Description)
		}
	}

	return bot.Respond(ctx, WithMessage(sb.String()))
}

func createCommand(
	commandname string,
	options []*discordgo.ApplicationCommandOption,
	description string,
) *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        commandname,
		Description: description,
		Options:     options,
	}
}

// retrieves the guild id from a given slash command context
func GuildID(ctx context.Context) string {
	i := ctx.Value(ctxKeyInteraction).(*discordgo.InteractionCreate)
	return i.GuildID
}
