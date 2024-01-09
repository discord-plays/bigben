package commands

import (
	"github.com/discord-plays/bigben/inter"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var _ CommandHandler = &debugCommand{}

var DebugCommands map[string]func()

type debugCommand struct {
	bot inter.MainBotInterface
}

func (x *debugCommand) Init(bot inter.MainBotInterface) {
	x.bot = bot
}

func (x *debugCommand) Command() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "debug",
		Description: "Debug options for Big Ben",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "cron", Description: "Force run cron function", Required: false, Choices: []discord.ApplicationCommandOptionChoiceString{
				{Name: "BingBong", Value: "bingBong"},
				{Name: "BingSetup", Value: "bingSetup"},
				{Name: "Christmas", Value: "christmas"},
				{Name: "New Year's", Value: "newYears"},
			}},
		},
	}
}

func (x *debugCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	if c := data.String("cron"); c != "" {
		if DebugCommands != nil {
			if f, ok := DebugCommands[c]; ok {
				f()
			}
		}
		_ = event.CreateMessage(discord.MessageCreate{
			Content: "Ran cron command",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
}
