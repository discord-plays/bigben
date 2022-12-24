package commands

import (
	"fmt"
	"github.com/MrMelon54/bigben/inter"
	"github.com/MrMelon54/bigben/tables"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/snabb/isoweek"
	"log"
	"strings"
	"time"
)

var _ CommandHandler = &leaderboardCommand{}

type leaderboardCommand struct {
	bot inter.MainBotInterface
}

type leaderboardCountTable struct {
	UserId snowflake.ID `xorm:"user_id"`
	Count  int64        `xorm:"a"`
}

type leaderboardAverageTable struct {
	UserId  snowflake.ID `xorm:"user_id"`
	Average float64      `xorm:"a"`
}

func (x *leaderboardCommand) Init(bot inter.MainBotInterface) {
	x.bot = bot
}

func (x *leaderboardCommand) Command() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:        "leaderboard",
		Description: "Show the leaderboard",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "time", Description: "Choose leaderboard time period", Required: true, Choices: []discord.ApplicationCommandOptionChoiceString{
				{Name: "Annually", Value: "annually"},
				{Name: "Bi-Annually", Value: "bi-annually"},
				{Name: "Quarterly", Value: "quarterly"},
				{Name: "Monthly", Value: "monthly"},
				{Name: "Weekly", Value: "weekly"},
				{Name: "Daily", Value: "daily"},
			}},
			discord.ApplicationCommandOptionString{Name: "type", Description: "Choose leaderboard type", Required: true, Choices: []discord.ApplicationCommandOptionChoiceString{
				{Name: "Total Clicks", Value: "total-clicks"},
				{Name: "Average Click Speed", Value: "average-speed"},
				{Name: "Slowest Click Speed", Value: "slowest-speed"},
				{Name: "Fastest Click Speed", Value: "fastest-speed"},
			}},
		},
	}
}

func (x *leaderboardCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	n := time.Now().UTC()
	var title string
	var rows []string

	// Send loading response
	err := event.DeferCreateMessage(false)
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to send interaction: %s\n", err)
	}

	var startTime time.Time
	var isDaily bool

	switch data.String("time") {
	case "annually":
		startTime = time.Date(n.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	case "bi-annually":
		if n.Month() < time.July {
			startTime = time.Date(n.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		} else {
			startTime = time.Date(n.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
		}
	case "quarterly":
		switch {
		case n.Month() < time.April:
			startTime = time.Date(n.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		case n.Month() < time.July:
			startTime = time.Date(n.Year(), time.April, 1, 0, 0, 0, 0, time.UTC)
		case n.Month() < time.October:
			startTime = time.Date(n.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
		default:
			startTime = time.Date(n.Year(), time.October, 1, 0, 0, 0, 0, time.UTC)
		}
	case "monthly":
		startTime = time.Date(n.Year(), n.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "weekly":
		y, w := n.ISOWeek()
		startTime = isoweek.StartTime(y, w, time.UTC)
	default:
		startTime = time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
		isDaily = true
	}

	startFlake := snowflake.New(startTime)

	// Figure out actual response
	switch data.String("type") {
	case "total-clicks":
		title = "Total Clicks Leaderboard"
		var a []leaderboardCountTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and msg_id > ? and won = ?", event.GuildID().String(), startFlake, true).GroupBy("user_id").OrderBy("a DESC, user_id DESC").Select("user_id, count(user_id) as a").Limit(10).Find(&a)
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
	case "average-speed":
		title = "Average Click Speed Leaderboard"
		var a []leaderboardAverageTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and msg_id > ? and won = ?", event.GuildID().String(), startFlake, true).GroupBy("user_id").OrderBy("a ASC, user_id DESC").Select("user_id, avg(speed) as a").Limit(10).Find(&a)
		if err != nil {
			log.Printf("[LeaderboardCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			duration, err := time.ParseDuration(fmt.Sprintf("%fms", j.Average))
			if err != nil {
				rows[i] = fmt.Sprintf("%d. <@%s> (%.0fs slowest reaction speed)", i+1, j.UserId, j.Average)
				return
			}
			duration = duration.Truncate(time.Millisecond)
			rows[i] = fmt.Sprintf("%d. <@%s> (%s average reaction speed)", i+1, j.UserId, duration)
		}
		if len(rows) == 0 {
			rows = []string{"No bong clicks found"}
		}
	case "slowest-speed":
		title = "Slowest Click Speed Leaderboard"
		var a []leaderboardAverageTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and msg_id > ? and won = ?", event.GuildID().String(), startFlake, true).GroupBy("user_id").OrderBy("a DESC, user_id DESC").Select("user_id, max(speed) as a").Limit(10).Find(&a)
		if err != nil {
			log.Printf("[LeaderboardCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			duration, err := time.ParseDuration(fmt.Sprintf("%fms", j.Average))
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
	case "fastest-speed":
		title = "Fastest Click Speed Leaderboard"
		var a []leaderboardAverageTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ? and msg_id > ? and won = ?", event.GuildID().String(), startFlake, true).GroupBy("user_id").OrderBy("a ASC, user_id DESC").Select("user_id, min(speed) as a").Limit(10).Find(&a)
		if err != nil {
			log.Printf("[LeaderbaordCommand] Database error: %s\n", err)
			return
		}
		rows = make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			duration, err := time.ParseDuration(fmt.Sprintf("%fms", j.Average))
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
	default:
		title = "Unknown Leaderboard"
		rows = []string{"Please pick a valid leaderboard type"}
	}
	if rows == nil {
		return
	}

	footerText := "Today"
	if !isDaily {
		footerText = "Since " + startTime.Format("2 Jan 2006")
	}

	updateBuilder := discord.NewMessageUpdateBuilder()
	updateBuilder.SetEmbeds(discord.Embed{
		Title:       title,
		Color:       0xd4af37,
		Description: strings.Join(rows, "\n"),
		Footer:      &discord.EmbedFooter{Text: footerText},
	})
	_, err = event.Client().Rest().UpdateInteractionResponse(event.ApplicationID(), event.Token(), updateBuilder.Build())
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to edit interaction: %s\n", err)
	}
}
