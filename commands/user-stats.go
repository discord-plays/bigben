package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/inter"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"log"
)

type userStatsCommand struct {
	bot inter.MainBotInterface
}

type userStatsTable struct {
	Count   int64   `xorm:"a"`
	Average float64 `xorm:"b"`
}

func (x *userStatsCommand) Init(bot inter.MainBotInterface) {
	x.bot = bot
}

func (x *userStatsCommand) Command() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "user-stats",
		Description: "Stats for a single user",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "user",
				Description: "User",
				Required:    true,
			},
		},
	}
}

func (x *userStatsCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	user := data.User("user")
	var a userStatsTable
	ok, err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and user_id = ?", event.GuildID().String(), user.ID.String()).Select("count(timestamp) as a, avg(time_to_sec(timestamp) - time_to_sec(message_timestamp)) as b").Get(&a)
	if err != nil {
		log.Printf("[UserStatsCommand] Database error: %s\n", err)
		return
	}
	if ok {
		_ = event.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Title: fmt.Sprintf("Stats for %s", user.String()),
					Color: 0xd4af37,
					Fields: []discord.EmbedField{
						{Name: "First Bong Count", Value: fmt.Sprint(a.Count)},
						{Name: "Average Reaction Time", Value: fmt.Sprintf("%.3fs", a.Average)},
					},
				},
			},
		})
		if err != nil {
			log.Printf("[UserStatsCommand] Failed to send interaction: %s\n", err)
		}
	} else {
		_ = event.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Title:       fmt.Sprintf("Stats for %s", user.String()),
					Color:       0xd4af37,
					Description: "No Sstats found",
				},
			},
		})
		if err != nil {
			log.Printf("[UserStatsCommand] Failed to send interaction: %s\n", err)
		}
	}
}
