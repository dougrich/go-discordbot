# go-discordbot

This is a bot built on top of bwmarrin's discordgo aimed at providing a more high level & ergonomic method for slash commands and simplifying a lot of the API.

Bots are composed of multiple packages:
```golang
func main() {
    discordbot.New(bottoken, config)
        .Include(...package...)
        .Include(...package...)
        .Start()
}
```

Each package is pretty simple:
```golang
type DiscordbotPackage interface {
    Name() string
    Register(*discordbot.Bot) (error)
}
```
Inside of the `Register` function, the package can repeatedly call `bot.AddCommand(name, options, description, handler)` which will automatically populate the help information & sort out how to respond to slash commands. Additionally, any background tasks here can be started and their cleanup can be connected with `bot.Defer`

For example, a simple add command see `go-discordbot/examples/cmd/add`: