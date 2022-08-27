package commands

import (
	"fmt"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

type setupCommand struct {
	bot utils.MainBotInterface
}

func (x *setupCommand) Init(bot utils.MainBotInterface) {
	x.bot = bot
}

func (x *setupCommand) Command() discordgo.ApplicationCommand {
	return discordgo.ApplicationCommand{
		Name:                     "setup",
		Description:              "Setup the bot",
		DefaultMemberPermissions: &manageServerPerm,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "channel",
				Description: "Choose the bong channel",
				Type:        discordgo.ApplicationCommandOptionChannel,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildText,
				},
			},
			{
				Name:        "role",
				Description: "Choose a bong role",
				Type:        discordgo.ApplicationCommandOptionRole,
			},
			{
				Name:        "emoji",
				Description: "Choose the bong emoji",
				Type:        discordgo.ApplicationCommandOptionString,
			},
		},
	}
}

func (x *setupCommand) Handler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildSettings, err := x.bot.GetGuildSettings(i.GuildID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to load guild settings",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	options := i.ApplicationCommandData().Options
	changed := false
	for _, j := range options {
		switch j.Name {
		case "channel":
			if guildSettings.BongWebhookId != "" {
				wh, err := s.WebhookDeleteWithToken(guildSettings.BongWebhookId, guildSettings.BongWebhookToken)
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("%#v\n", wh)
			}
			guildSettings.BongChannelId = j.ChannelValue(s).ID
			webhook, err := s.WebhookCreate(guildSettings.BongChannelId, "Big Ben", "")
			if err != nil {
				return
			}
			guildSettings.BongWebhookId = webhook.ID
			guildSettings.BongWebhookToken = webhook.Token
			changed = true
		case "role":
			guildSettings.BongRoleId = j.RoleValue(s, i.GuildID).ID
			changed = true
		case "emoji":
			strVal := j.StringValue()
			guildSettings.BongEmoji = strings.Join(utils.DecodeAllDiscordEmoji(strVal), "")
			changed = true
		}
	}
	chanVal := "None"
	roleVal := "None"
	emojiVal := "None"
	if guildSettings.BongChannelId != "" {
		chanVal = fmt.Sprintf("<#%s>", guildSettings.BongChannelId)
	}
	if guildSettings.BongRoleId != "" {
		roleVal = fmt.Sprintf("<@&%s>", guildSettings.BongRoleId)
	}
	if guildSettings.BongEmoji != "" {
		emojiVal = guildSettings.BongEmoji
	}
	if changed {
		err = x.bot.PutGuildSettings(guildSettings)
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Failed to save guild settings",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Setup Big Ben",
					Color: 0xd4af37,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Channel",
							Value: chanVal,
						},
						{
							Name:  "Role",
							Value: roleVal,
						},
						{
							Name:  "Emoji",
							Value: emojiVal,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Made by MrMelon54",
					},
				},
			},
			Components: []discordgo.MessageComponent{},
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}
