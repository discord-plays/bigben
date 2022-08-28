package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
	"log"
)

type userStatsCommand struct {
	bot utils.MainBotInterface
}

type userStatsTable struct {
	Count   int64   `xorm:"a"`
	Average float64 `xorm:"b"`
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
	var a userStatsTable
	ok, err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and user_id = ?", i.GuildID, user.ID).Select("count(timestamp) as a, avg(time_to_sec(timestamp) - time_to_sec(message_timestamp)) as b").Get(&a)
	if err != nil {
		log.Printf("[UserStatsCommand] Database error: %s\n", err)
		return
	}
	if ok {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title: fmt.Sprintf("Stats for %s", user.String()),
						Color: 0xd4af37,
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:  "First Bong Count",
								Value: fmt.Sprint(a.Count),
							},
							{
								Name:  "Average Reaction Time",
								Value: fmt.Sprintf("%.3fs", a.Average),
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Printf("[UserStatsCommand] Failed to send interaction: %s\n", err)
		}
	} else {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       fmt.Sprintf("Stats for %s", user.String()),
						Color:       0xd4af37,
						Description: "No stats found",
					},
				},
			},
		})
		if err != nil {
			log.Printf("[UserStatsCommand] Failed to send interaction: %s\n", err)
		}
	}
}
