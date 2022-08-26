package commands

import (
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
)

type userStatsCommand struct{}

func (x *userStatsCommand) Init(utils.MainBotInterface) {}

func (x *userStatsCommand) Command() discordgo.ApplicationCommand {
	return discordgo.ApplicationCommand{
		Name:        "user-stats",
		Description: "Stats for a single user",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user-option",
				Description: "User",
				Required:    true,
			},
		},
	}
}

func (x *userStatsCommand) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "Running user stats command for " + user.String()},
	})
}
