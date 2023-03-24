package commands

import (
	"fmt"
	"github.com/MrMelon54/bigben/inter"
	"github.com/MrMelon54/bigben/tables"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"log"
	"time"
)

var _ CommandHandler = &userStatsCommand{}

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
	ok, err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and user_id = ? and won = 1", event.GuildID().String(), user.ID.String()).Select("count(speed) as a, avg(speed) as b").Get(&a)
	if err != nil {
		log.Printf("[UserStatsCommand] Database error: %s\n", err)
		return
	}
	if ok {
		avg := time.Duration(int64(a.Average * float64(time.Millisecond))).Round(time.Millisecond)
		_ = event.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Title: fmt.Sprintf("Stats for %s", user.Username),
					Color: 0xd4af37,
					Fields: []discord.EmbedField{
						{Name: "User", Value: user.String()},
						{Name: "Total bong count", Value: fmt.Sprint(a.Count)},
						{Name: "Average reaction time", Value: avg.String()},
					},
				},
			},
			AllowedMentions: &discord.AllowedMentions{
				//Parse: []discord.AllowedMentionType{discord.AllowedMentionTypeUsers},
				Users: []snowflake.ID{user.ID},
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
