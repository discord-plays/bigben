package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/discord-plays/bigben/database"
	"github.com/discord-plays/bigben/inter"
	"github.com/discord-plays/bigben/logger"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"time"
)

var _ CommandHandler = &userStatsCommand{}

type userStatsCommand struct {
	bot inter.MainBotInterface
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
	stats, err := x.bot.Engine().UserStats(context.Background(), database.UserStatsParams{GuildID: *event.GuildID(), UserID: user.ID})
	noRows := errors.Is(err, sql.ErrNoRows)
	if err != nil && !noRows {
		logger.Logger.Error("UserStats", "err", err)
		return
	}

	if noRows {
		err = event.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Title:       fmt.Sprintf("Stats for %s", user.String()),
					Color:       0xd4af37,
					Description: "No Sstats found",
				},
			},
		})
	} else {
		avg := time.Duration(int64(stats.AverageSpeed * float64(time.Millisecond))).Round(time.Millisecond)
		err = event.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{
				{
					Title: fmt.Sprintf("Stats for %s", user.Username),
					Color: 0xd4af37,
					Fields: []discord.EmbedField{
						{Name: "User", Value: user.String()},
						{Name: "Total bong count", Value: fmt.Sprint(stats.TotalBongs)},
						{Name: "Average reaction time", Value: avg.String()},
					},
				},
			},
			AllowedMentions: &discord.AllowedMentions{
				Users: []snowflake.ID{user.ID},
			},
		})
	}
	if err != nil {
		logger.Logger.Error("Failed to send interaction", "err", err)
	}
}
