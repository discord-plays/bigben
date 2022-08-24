package commands

import "github.com/bwmarrin/discordgo"

var (
	prebuiltHandlers = []CommandHandler{
		&setupCommand{},
		&leaderboardCommand{},
		&userStatsCommand{},
	}
)

type CommandList []*discordgo.ApplicationCommand
type CommandHandler interface {
	Command() discordgo.ApplicationCommand
	Handler(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func InitCommands() (CommandList, map[string]CommandHandler) {
	var commands CommandList
	commandHandlers := make(map[string]CommandHandler, len(prebuiltHandlers))

	for _, i := range prebuiltHandlers {
		c := i.Command()
		commands = append(commands, &c)
		commandHandlers[c.Name] = i
	}
	return commands, commandHandlers
}
