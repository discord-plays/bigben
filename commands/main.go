package commands

import (
	"github.com/MrMelon54/bigben/inter"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"os"
)

var prebuiltHandlers = []CommandHandler{
	&setupCommand{},
	&leaderboardCommand{},
	&userStatsCommand{},
}

type CommandList []discord.ApplicationCommandCreate
type CommandHandler interface {
	Init(botInterface inter.MainBotInterface)
	Command() discord.SlashCommandCreate
	Handler(event *events.ApplicationCommandInteractionCreate)
}

func InitCommands(bot inter.MainBotInterface) (CommandList, map[string]CommandHandler) {
	var commands CommandList
	commandHandlers := make(map[string]CommandHandler, len(prebuiltHandlers))

	for _, i := range prebuiltHandlers {
		i.Init(bot)
		c := i.Command()
		commands = append(commands, &c)
		commandHandlers[c.Name] = i
	}
	if os.Getenv("DEBUG_MODE") == "1" {
		i := &debugCommand{}
		i.Init(bot)
		c := i.Command()
		commands = append(commands, &c)
		commandHandlers[c.Name] = i
	}
	return commands, commandHandlers
}
