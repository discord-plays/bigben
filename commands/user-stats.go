package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
)

type userStatsCommand struct {
	bot utils.MainBotInterface
}

func (x *userStatsCommand) Init(bot utils.MainBotInterface) {
	x.bot = bot
}

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
	var a []tables.BongLog
	err := x.bot.Engine().Where("guild_id = ? and user_id = ?", i.GuildID, user.ID).Find(&a)
	if err != nil {
		return
	}
	var total float64
	var count int64
	for _, i := range a {
		total += i.Timestamp.Sub(i.Timestamp).Seconds()
		count++
	}
	average := total / float64(count)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: fmt.Sprintf("Stats for %s", user.String()),
					Color: 0xd4af37,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "First Bong Count",
							Value: fmt.Sprint(count),
						},
						{
							Name:  "Average Reaction Time",
							Value: fmt.Sprintf("%.3f", average),
						},
					},
				},
			},
		},
	})
}
