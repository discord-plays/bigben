package commands

import (
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
)

var (
	manageServerPerm int64 = discordgo.PermissionManageServer
	prebuiltHandlers       = []CommandHandler{
		&setupCommand{},
		&leaderboardCommand{},
		&userStatsCommand{},
	}
)

type CommandList []*discordgo.ApplicationCommand
type CommandHandler interface {
	Init(utils.MainBotInterface)
	Command() discordgo.ApplicationCommand
	Handler(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func InitCommands(bot utils.MainBotInterface) (CommandList, map[string]CommandHandler) {
	var commands CommandList
	commandHandlers := make(map[string]CommandHandler, len(prebuiltHandlers))

	for _, i := range prebuiltHandlers {
		i.Init(bot)
		c := i.Command()
		commands = append(commands, &c)
		commandHandlers[c.Name] = i
	}
	return commands, commandHandlers
}
