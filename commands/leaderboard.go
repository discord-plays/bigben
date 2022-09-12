package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/inter"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"log"
	"os"
	"strings"
	"time"
)

type leaderboardCommand struct {
	bot inter.MainBotInterface
}

type leaderboardCountTable struct {
	UserId string `xorm:"user_id"`
	Count  int64  `xorm:"a"`
}

type leaderboardAverageTable struct {
	UserId  string  `xorm:"user_id"`
	Average float64 `xorm:"a"`
}

func (x *leaderboardCommand) Init(bot inter.MainBotInterface) {
	x.bot = bot
}

func (x *leaderboardCommand) Command() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "leaderboard",
		Description: "Show the leaderboard",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{Name: "click-total", Description: "Click total leaderboard"},
			discord.ApplicationCommandOptionSubCommand{Name: "click-speed", Description: "Click speed leaderboard"},
			discord.ApplicationCommandOptionSubCommand{Name: "click-long", Description: "Longest click leaderboard"},
			discord.ApplicationCommandOptionSubCommand{Name: "click-short", Description: "Shortest click leaderboard"},
		},
	}
}

func (x *leaderboardCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	var title string
	var rows []string

	// Send loading response
	myBuilder := discord.NewMessageCreateBuilder()
	myBuilder.SetEmbeds(discord.Embed{Title: "Leaderboards will be unavailable until further notice", Color: 0xd4af37})
	err := event.CreateMessage(myBuilder.Build())
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to send interaction: %s\n", err)
	}

	if os.Getenv("THIS_IS_INVALID") != "THIS_IS_INVALID" {
		return
	}

	// Send loading response
	err = event.DeferCreateMessage(false)
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to send interaction: %s\n", err)
	}

	// Figure out actual response
	switch *data.SubCommandName {
	case "click-total":
		title = "Click Total Leaderboard"
		var a []leaderboardCountTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ?", event.GuildID().String()).GroupBy("user_id").OrderBy("a DESC, user_id DESC").Select("user_id, count(user_id) as a").Find(&a)
		if err != nil {
			log.Printf("[LeaderboardCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			rows[i] = fmt.Sprintf("%d. <@%s> (%d bongs)", i+1, j.UserId, j.Count)
		}
		if len(rows) == 0 {
			rows = []string{"No bong clicks found"}
		}
	case "click-speed":
		title = "Click Speed Leaderboard"
		var a []leaderboardAverageTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ?", event.GuildID().String()).GroupBy("user_id").OrderBy("a ASC, user_id DESC").Select("user_id, avg(time_to_sec(timestamp) - time_to_sec(message_timestamp)) as a").Find(&a)
		if err != nil {
			log.Printf("[LeaderboardCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			rows[i] = fmt.Sprintf("%d. <@%s> (%.3fs average reaction speed)", i+1, j.UserId, j.Average)
		}
		if len(rows) == 0 {
			rows = []string{"No bong clicks found"}
		}
	case "click-long":
		title = "Click Long Leaderboard"
		var a []leaderboardAverageTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ?", event.GuildID().String()).GroupBy("user_id").OrderBy("a DESC, user_id DESC").Select("user_id, max(time_to_sec(timestamp) - time_to_sec(message_timestamp)) as a").Find(&a)
		if err != nil {
			log.Printf("[LeaderboardCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			duration, err := time.ParseDuration(fmt.Sprintf("%fs", j.Average))
			if err != nil {
				rows[i] = fmt.Sprintf("%d. <@%s> (%.0fs slowest reaction speed)", i+1, j.UserId, j.Average)
				return
			}
			duration = duration.Truncate(time.Millisecond)
			rows[i] = fmt.Sprintf("%d. <@%s> (%s slowest reaction speed)", i+1, j.UserId, duration)
		}
		if len(rows) == 0 {
			rows = []string{"No bong clicks found"}
		}
	case "click-short":
		title = "Click Short Leaderboard"
		var a []leaderboardAverageTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ?", event.GuildID().String()).GroupBy("user_id").OrderBy("a ASC, user_id DESC").Select("user_id, min(time_to_sec(timestamp) - time_to_sec(message_timestamp)) as a").Find(&a)
		if err != nil {
			log.Printf("[LeaderbaordCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			duration, err := time.ParseDuration(fmt.Sprintf("%fs", j.Average))
			if err != nil {
				rows[i] = fmt.Sprintf("%d. <@%s> (%.0fs quickest reaction speed)", i+1, j.UserId, j.Average)
				return
			}
			duration = duration.Truncate(time.Millisecond)
			rows[i] = fmt.Sprintf("%d. <@%s> (%s quickest reaction speed)", i+1, j.UserId, duration)
		}
		if len(rows) == 0 {
			rows = []string{"No bong clicks found"}
		}
	}
	if rows == nil {
		return
	}
	updateBuilder := discord.NewMessageCreateBuilder()
	updateBuilder.SetEmbeds(discord.Embed{
		Title:       title,
		Color:       0xd4af37,
		Description: strings.Join(rows, "\n"),
	})
	err = event.CreateMessage(updateBuilder.Build())
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to edit interaction: %s\n", err)
	}
}
