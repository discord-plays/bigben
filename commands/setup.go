package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/discord-plays/bigben/assets"
	"github.com/discord-plays/bigben/database/types"
	"github.com/discord-plays/bigben/inter"
	"github.com/discord-plays/bigben/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/json"
	"log"
	"strings"
	"time"
)

var _ CommandHandler = &setupCommand{}

type setupCommand struct {
	bot inter.MainBotInterface
}

func (x *setupCommand) Init(bot inter.MainBotInterface) {
	x.bot = bot
}

func (x *setupCommand) Command() discord.SlashCommandCreate {
	return discord.SlashCommandCreate{
		Name:                     "setup",
		Description:              "Setup the bot",
		DefaultMemberPermissions: json.NewNullablePtr[discord.Permissions](discord.PermissionManageGuild),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{
				Name:        "channel",
				Description: "Choose the bong channel",
				Required:    true,
				ChannelTypes: []discord.ChannelType{
					discord.ChannelTypeGuildText,
				},
			},
			discord.ApplicationCommandOptionRole{
				Name:        "role",
				Description: "Choose a bong role",
				Required:    true,
			},
			discord.ApplicationCommandOptionString{
				Name:        "emoji",
				Description: "Choose the bong emoji",
				Required:    true,
			},
		},
	}
}

func (x *setupCommand) Handler(event *events.ApplicationCommandInteractionCreate) {
	if event.GuildID() == nil {
		return
	}
	guildSettings, err := x.bot.Engine().GetGuild(context.Background(), *event.GuildID())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = event.CreateMessage(discord.MessageCreate{
			Content: "Failed to load guild settings",
			Flags:   discord.MessageFlagEphemeral,
		})
		log.Printf("Failed to load guild settings: %s\n", err)
		return
	}
	guildSettings.ID = *event.GuildID()
	data := event.SlashCommandInteractionData()
	changed := false
	for _, j := range data.Options {
		switch j.Name {
		case "channel":
			guildSettings.BongChannelID = data.Channel("channel").ID
			var create bool
			if guildSettings.BongWebhookID != 0 {
				getWebhook, err := x.bot.Session().Rest().GetWebhook(guildSettings.BongWebhookID)
				if err != nil || getWebhook == nil {
					create = true
				}
			} else {
				create = true
			}
			var wh discord.Webhook
			var token string
			if create {
				n := utils.GetStartOfHourTime().Add(time.Hour)
				a, err := x.bot.Session().Rest().CreateWebhook(guildSettings.BongChannelID, discord.WebhookCreate{
					Name:   "Big Ben",
					Avatar: assets.ReadClockFaceByTimeAsOptionalIcon(n),
				})
				if err != nil {
					continue
				}
				token = a.Token
				wh = a
			} else {
				wh, err = x.bot.Session().Rest().UpdateWebhook(guildSettings.BongWebhookID, discord.WebhookUpdate{
					Name: utils.PString("Big Ben"),
				})
				if err != nil {
					continue
				}
			}
			guildSettings.BongWebhookID = wh.ID()
			if token != "" {
				guildSettings.BongWebhookToken = token
			}
			changed = true
		case "role":
			guildSettings.BongRoleID = data.Role("role").ID
			changed = true
		case "emoji":
			strVal := data.String("emoji")
			guildSettings.BongEmoji = strings.Join(utils.DecodeAllDiscordEmoji(strVal), "")
			changed = true
		}
	}
	chanVal := "None"
	roleVal := "None"
	emojiVal := "None"
	if guildSettings.BongChannelID != 0 {
		chanVal = fmt.Sprintf("<#%s>", guildSettings.BongChannelID)
	}
	if guildSettings.BongRoleID != 0 {
		roleVal = fmt.Sprintf("<@&%s>", guildSettings.BongRoleID)
	}
	if guildSettings.BongEmoji != "" {
		emojiVal = guildSettings.BongEmoji
	}
	if changed {
		err = x.bot.Engine().UpdateGuild(context.Background(), types.ConvertGuildToParams(guildSettings))
		if err != nil {
			_ = event.CreateMessage(discord.MessageCreate{
				Content: "Failed to save guild settings",
				Flags:   discord.MessageFlagEphemeral,
			})
			log.Printf("Failed to save guild settings: %s\n", err)
			return
		}
	}
	_ = event.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Title: "Setup Big Ben",
				Color: 0xd4af37,
				Fields: []discord.EmbedField{
					{Name: "Channel", Value: chanVal},
					{Name: "Role", Value: roleVal},
					{Name: "Emoji", Value: emojiVal},
				},
				Footer: &discord.EmbedFooter{
					Text:         "Made by MrMelon54",
					IconURL:      "",
					ProxyIconURL: "",
				},
			},
		},
		Components: []discord.ContainerComponent{},
		Flags:      discord.MessageFlagEphemeral,
	})
}
