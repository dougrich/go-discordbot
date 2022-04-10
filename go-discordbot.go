package discordbot

import (
	"context"
	"fmt"
	"io"
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
	if !isSubcommand {
		return true
	}

	for _, o := range c.cmd.Options {
		if o.Name == d.Options[0].Name {
			return true
		}
	}
	return false
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
				if o.Type == discordgo.ApplicationCommandOptionSubCommand && o.Name == cmd.Options[i].Name {
					bot.ErrRegistration = fmt.Errorf("Failed to register \"/%s %s\", subcommand is already registered", c.Name, cmd.Options[i].Name)
					return
				}
			}
			c.Options = append(c.Options, cmd.Options[i])
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
		if err := m.ModifyInteraction(bot, response); err != nil {
			return err
		}
	}
	err := bot.session.InteractionRespond(i.Interaction, response)
	if err != nil {
		return err
	}
	msgAfter, err := bot.session.InteractionResponse(bot.session.State.User.ID, i.Interaction)
	if err != nil {
		return err
	}
	for _, m := range mods {
		if err := m.ModifyMessageAfter(bot, msgAfter); err != nil {
			return err
		}
	}
	return nil
}

func (bot *Bot) Message(channelID string, mods ...MessageModifier) error {
	msgBefore := &discordgo.MessageSend{
		TTS: false,
	}
	for _, m := range mods {
		if err := m.ModifyMessage(bot, msgBefore); err != nil {
			return err
		}
	}
	msgAfter, err := bot.session.ChannelMessageSendComplex(channelID, msgBefore)
	if err != nil {
		return err
	}
	for _, m := range mods {
		if err := m.ModifyMessageAfter(bot, msgAfter); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) Include(p BotPackage) *Bot {
	if b.ErrRegistration == nil {
		err := p.Register(b)
		if err != nil {
			b.ErrRegistration = &errRegistration{p.Name(), err}
		} else if b.ErrRegistration != nil {
			b.ErrRegistration = &errRegistration{p.Name(), b.ErrRegistration}
		}
	}
	return b
}

func (bot *Bot) Start() {
	if bot.ErrRegistration != nil {
		log.Panicf("discordbot: invalid registration:\n%v", bot.ErrRegistration)
	}

	defer bot.runDefers()

	log.Printf("discordbot: starting bot %s", bot.options.Token)
	if bot.options.GuildID != "" {
		log.Printf("discordbot: >> scoped to %s", bot.options.GuildID)
	}

	s, err := discordgo.New("Bot " + bot.options.Token)
	if err != nil {
		log.Panicf("discordbot: invalid bot parameters: %v", err)
	}
	bot.session = s

	err = s.Open()
	if err != nil {
		log.Panicf("discordbot: unable to start discord session: %v", err)
	}
	defer s.Close()
	registeredCommands := make([]*discordgo.ApplicationCommand, len(bot.commands))
	for i, v := range bot.commands {
		log.Printf("discordbot: trying to create %v as %s", v, s.State.User.ID)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, bot.options.GuildID, v)
		if err != nil {
			log.Panicf("discordbot: cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// register callbacks
	s.AddHandler(bot.handleInteraction)

	// wait until we're done
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("discordbot: starting shutdown from signal")

	// optionally remove commands
	if bot.options.RemoveCommands {
		log.Println("discordbot: removing commands...")
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(s.State.User.ID, bot.options.GuildID, v.ID)
			if err != nil {
				log.Panicf("discordbot: cannot delete '%v' command: %v", v.Name, err)
			}
		}
		log.Println("discordbot: commands removed")
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
		log.Printf("discordbot: command %s", matcher.commandname)
		err := matcher.handler(ctx, arguments)
		if err != nil {
			log.Printf("discordbot: error occured in %s: %v", matcher.commandname, err)
		}
		return
	}
	// unmatched
	log.Print("discordbot: unmatched command")
}

func (bot *Bot) helpInteraction(ctx context.Context, _ *Arguments) error {
	var sb strings.Builder
	for _, m := range bot.matchers {
		var options []*discordgo.ApplicationCommandOption
		for _, o := range m.options {
			if o.Type == discordgo.ApplicationCommandOptionSubCommand {
				printHelp(&sb, m.commandname, o.Name, o.Options, o.Description)
			} else {
				options = append(options, o)
			}
		}
		if len(options) > 0 {
			printHelp(&sb, m.commandname, "", options, m.description)
		}
	}

	return bot.Respond(ctx, WithMessage(sb.String()))
}

func printHelp(out io.Writer, commandname string, subcommand string, options []*discordgo.ApplicationCommandOption, description string) {
	fmt.Fprintf(out, "\n**/%s", commandname)
	if subcommand != "" {
		fmt.Fprintf(out, " %s", subcommand)
	}
	fmt.Fprintf(out, "**_")
	for _, o := range options {
		fmt.Fprintf(out, " %s:%s", o.Name, o.Type.String())
	}
	fmt.Fprintf(out, "_\n\t%s\n", description)
	for _, o := range options {
		fmt.Fprintf(out, "\t*%s:%s*\n\t\t%s\n", o.Name, o.Type.String(), o.Description)
	}
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
