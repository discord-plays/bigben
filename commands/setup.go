package commands

import (
	"fmt"
	"github.com/Succo/emoji"
	"github.com/bwmarrin/discordgo"
	"strings"
)

var manageServerPerm int64 = discordgo.PermissionManageServer

type setupCommand struct{}

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
	options := i.ApplicationCommandData().Options
	for _, j := range options {
		switch j.Name {
		case "channel":
			fmt.Printf("Using channel: %s\n", j.ChannelValue(s).Name)
		case "role":
			fmt.Printf("Using role: %s\n", j.RoleValue(s, i.GuildID).Name)
		case "emoji":
			emojiStr := emoji.FindString(j.StringValue(), -1)
			fmt.Printf("Using emoji: '%s'\n", strings.Join(emojiStr, ""))
		}
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Running setup command",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
