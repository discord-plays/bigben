package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

type leaderboardCommand struct {
	bot utils.MainBotInterface
}

type leaderboardCountTable struct {
	UserId string `xorm:"user_id"`
	Count  int64  `xorm:"a"`
}

type leaderboardAverageTable struct {
	UserId  string  `xorm:"user_id"`
	Average float64 `xorm:"a"`
}

func (x *leaderboardCommand) Init(bot utils.MainBotInterface) {
	x.bot = bot
}

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
	var title string
	var rows []string

	// Send loading response
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Loading leaderboard...",
					Color: 0xd4af37,
				},
			},
		},
	})
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to send interaction: %s\n", err)
	}

	// Figure out actual response
	switch options[0].Name {
	case "click-total":
		title = "Click Total Leaderboard"
		var a []leaderboardCountTable
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ?", i.GuildID).GroupBy("user_id").OrderBy("a DESC, user_id DESC").Select("user_id, count(user_id) as a").Find(&a)
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
		err := x.bot.Engine().Table(&tables.BongLog{}).Where("guild_id = ?", i.GuildID).GroupBy("user_id").OrderBy("a DESC, user_id DESC").Select("user_id, avg(time_to_sec(timestamp) - time_to_sec(message_timestamp)) as a").Find(&a)
		if err != nil {
			log.Printf("[LeaderboardCommand] Database error: %s\n", err)
			return
		}
		rows := make([]string, len(a))
		for i, j := range a {
			if i >= 10 {
				break
			}
			rows[i] = fmt.Sprintf("%d. <@%s> (%.3f bongs)", i+1, j.UserId, j.Average)
		}
		if len(rows) == 0 {
			rows = []string{"No bong clicks found"}
		}
	}
	if rows == nil {
		return
	}
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{
			{
				Title:       title,
				Color:       0xd4af37,
				Description: strings.Join(rows, "\n"),
			},
		},
	})
	if err != nil {
		log.Printf("[LeaderboardCommand] Failed to send interaction: %s\n", err)
	}
}
