package commands

import (
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
)

type leaderboardCommand struct{}

func (x *leaderboardCommand) Init(utils.MainBotInterface) {}

func (x *leaderboardCommand) Command() discordgo.ApplicationCommand {
	return discordgo.ApplicationCommand{
		Name:        "leaderboard",
		Description: "Show the leaderboard",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "click-total",
				Description: "Click total leaderboard",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "click-speed",
				Description: "Click speed leaderboard",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}
}

func (x *leaderboardCommand) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	var content string
	switch options[0].Name {
	case "click-total":
		content = "Click total"
	case "click-speed":
		content = "Click speed"
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
}
