package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/assets"
	"github.com/MrMelon54/BigBen/inter"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/snowflake/v2"
	"strings"
	"time"
)

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
		DefaultMemberPermissions: json.NewOptional[discord.Permissions](discord.PermissionManageServer),
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
	guildSettings, err := x.bot.GetGuildSettings(*event.GuildID())
	if err != nil {
		_ = event.CreateMessage(discord.MessageCreate{
			Content: "Failed to load guild settings",
			Flags:   discord.MessageFlagEphemeral,
		})
		return
	}
	data := event.SlashCommandInteractionData()
	changed := false
	for _, j := range data.Options {
		switch j.Name {
		case "channel":
			guildSettings.BongChannelId = utils.XormSnowflake(data.Channel("channel").ID)
			var create bool
			if guildSettings.BongWebhookId != 0 {
				getWebhook, err := x.bot.Session().Rest().GetWebhook(snowflake.ID(guildSettings.BongWebhookId))
				if err != nil || getWebhook == nil {
					create = true
				}
			}
			var wh discord.Webhook
			var token string
			if create {
				n := utils.GetStartOfHourTime().Add(time.Hour)
				a, err := x.bot.Session().Rest().CreateWebhook(snowflake.ID(guildSettings.BongChannelId), discord.WebhookCreate{
					Name:   "Big Ben",
					Avatar: assets.ReadClockFaceByTimeAsOptionalIcon(n),
				})
				if err != nil {
					continue
				}
				token = a.Token
				wh = a
			} else {
				wh, err = x.bot.Session().Rest().UpdateWebhook(snowflake.ID(guildSettings.BongWebhookId), discord.WebhookUpdate{
					Name: utils.PString("Big Ben"),
				})
				if err != nil {
					continue
				}
			}
			guildSettings.BongWebhookId = utils.XormSnowflake(wh.ID())
			if token != "" {
				guildSettings.BongWebhookToken = token
			}
			changed = true
		case "role":
			guildSettings.BongRoleId = utils.XormSnowflake(data.Role("role").ID)
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
	if guildSettings.BongChannelId != 0 {
		chanVal = fmt.Sprintf("<#%s>", snowflake.ID(guildSettings.BongChannelId))
	}
	if guildSettings.BongRoleId != 0 {
		roleVal = fmt.Sprintf("<@&%s>", snowflake.ID(guildSettings.BongRoleId))
	}
	if guildSettings.BongEmoji != "" {
		emojiVal = guildSettings.BongEmoji
	}
	if changed {
		err = x.bot.PutGuildSettings(guildSettings)
		if err != nil {
			_ = event.CreateMessage(discord.MessageCreate{
				Content: "Failed to save guild settings",
				Flags:   discord.MessageFlagEphemeral,
			})
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
