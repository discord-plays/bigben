package main

import (
	"fmt"
	"github.com/MrMelon54/BigBen/commands"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
	"log"
	"sync"
	"time"
	"xorm.io/xorm"
)

type BigBen struct {
	engine          *xorm.Engine
	appId           string
	guildId         string
	session         *discordgo.Session
	commands        commands.CommandList
	commandHandlers map[string]commands.CommandHandler
	bongLock        *sync.Mutex
	currentBong     *utils.CurrentBong
	cron            *cron.Cron
}

func (b *BigBen) AppId() string { return b.appId }

func (b *BigBen) GuildId() string { return b.guildId }

func (b *BigBen) Session() *discordgo.Session { return b.session }
func NewBigBen(engine *xorm.Engine, token, appId, guildId string) (*BigBen, error) {
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	bot.AddHandler(func(s *discordgo.Session, _ *discordgo.Ready) {
		_ = s.UpdateGameStatus(0, "Bong")
	})
	bot.Identify.Intents = discordgo.IntentGuilds | discordgo.IntentGuildMessages
	err = bot.Open()
	if err != nil {
		return nil, err
	}
	return (&BigBen{}).init(engine, appId, guildId, bot)
}
func (b *BigBen) init(engine *xorm.Engine, appId, guildId string, bot *discordgo.Session) (*BigBen, error) {
	b.engine = engine
	b.appId = appId
	b.guildId = guildId
	b.session = bot
	b.commands, b.commandHandlers = commands.InitCommands(b)
	b.bongLock = &sync.Mutex{}
	b.currentBong = nil
	err := b.updateCommands()
	if err != nil {
		return nil, err
	}
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := b.commandHandlers[i.ApplicationCommandData().Name]; ok {
				h.Handler(s, i)
			}
		case discordgo.InteractionMessageComponent:
			d := i.MessageComponentData()
			if d.ComponentType == discordgo.ButtonComponent {
				if d.CustomID == "bong" {
					b.ClickBong(s, i)
				}
			}
		}
	})
	b.cron = cron.New()
	_ = b.cron.AddFunc("0 0 * * * *", b.bingBong)
	_ = b.cron.AddFunc("*/2 * * * * *", b.updateMessageData)
	b.cron.Start()
	return b, nil
}

func (b *BigBen) Exit() error {
	return b.session.Close()
}

func (b *BigBen) bingBong() {
	log.Println("[bingBong()] Sending hourly bong")
	all, err := b.GetAllGuildSettings()
	if err != nil {
		log.Printf("[bingBong()] Error: %s\n", err)
		return
	}
	b.bongLock.Lock()
	now := utils.GetStartOfHourTime()
	title := utils.GetBongTitle(now)
	sTime := title.T
	eTime := title.T.Add(time.Hour * 24)
	if b.currentBong != nil {
		b.currentBong.Kill()
	}
	b.currentBong = utils.NewCurrentBong(b.engine, title.S, sTime, eTime)
	b.currentBong.RandomGuildData(all)
	for _, i := range all {
		g := b.currentBong.GuildMapItem(i.GuildId)
		g.Lock.RLock()
		go b.internalSendBongMessage(g, i.BongChannelId, b.currentBong.Text, utils.ConvertToComponentEmoji(g.Emoji))
		g.Lock.RUnlock()
	}
	b.bongLock.Unlock()
}

func (b *BigBen) updateMessageData() {
	if b.currentBong == nil {
		return
	}
	all, err := b.GetAllGuildSettings()
	if err != nil {
		log.Printf("[updateMessageData()] Error: %s\n", err)
		return
	}
	b.bongLock.Lock()
	for _, i := range all {
		g := b.currentBong.GuildMapItem(i.GuildId)
		if g == nil {
			continue
		}
		g.Lock.Lock()
		if g.Dirty {
			go b.internalEditBongMessage(i.BongChannelId, g.MessageId, b.currentBong.Text, utils.ConvertToComponentEmoji(g.Emoji), g.ClickNames)
		}
		g.Lock.Unlock()
	}
	b.bongLock.Unlock()
}

func (b *BigBen) updateCommands() error {
	_, err := b.session.ApplicationCommandBulkOverwrite(b.appId, b.guildId, b.commands)
	if err != nil {
		return fmt.Errorf("bulk overwrite application commands error: %s", err)
	}
	return nil
}

func (b *BigBen) GetAllGuildSettings() ([]tables.GuildSettings, error) {
	var g []tables.GuildSettings
	err := b.engine.Find(&g)
	return g, err
}

func (b *BigBen) GetGuildSettings(guildId string) (tables.GuildSettings, error) {
	var g tables.GuildSettings
	_, err := b.engine.Where("guild_id = ?", guildId).Get(&g)
	g.GuildId = guildId
	return g, err
}

func (b *BigBen) PutGuildSettings(guildSettings tables.GuildSettings) error {
	ok, err := b.engine.Where("guild_id = ?", guildSettings.GuildId).Update(&guildSettings)
	if err != nil {
		return err
	}
	if ok == 0 {
		_, err = b.engine.Insert(&guildSettings)
		return err
	}
	return nil
}

func (b *BigBen) internalSendBongMessage(g *utils.GuildCurrentBong, channelId, title string, emoji discordgo.ComponentEmoji) {
	m, err := b.session.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Embeds:     b.bongEmbeds(title),
		Components: b.bongComponents(emoji, []string{}),
	})
	if err != nil {
		log.Printf("[internalSendBongMessage(\"%s\")] Error: %s\n", channelId, err)
	}
	g.Lock.Lock()
	g.MessageId = m.ID
	g.MessageTime = m.Timestamp
	g.Lock.Unlock()
}

func (b *BigBen) internalEditBongMessage(channelId, messageId, title string, emoji discordgo.ComponentEmoji, names []string) {
	if messageId == "" {
		return
	}
	_, err := b.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Embeds:     b.bongEmbeds(title),
		Components: b.bongComponents(emoji, names),
		Channel:    channelId,
		ID:         messageId,
	})
	if err != nil {
		log.Printf("[internalEditBongMessage(\"%s\")] Error: %s\n", channelId, err)
	}
}

func (b *BigBen) bongEmbeds(title string) []*discordgo.MessageEmbed {
	return []*discordgo.MessageEmbed{
		{
			Title: title,
			Color: 0xd4af37,
		},
	}
}

func (b *BigBen) bongComponents(emoji discordgo.ComponentEmoji, names []string) []discordgo.MessageComponent {
	if len(names) > 3 {
		names = names[:3]
	}
	l := ""
	if len(names) == 1 {
		l = "1 click"
	} else if len(names) > 1 {
		l = fmt.Sprintf("%d clicks", len(names))
	}
	a := []discordgo.MessageComponent{
		discordgo.Button{
			Label:    l,
			Style:    discordgo.SecondaryButton,
			Disabled: false,
			Emoji:    emoji,
			CustomID: "bong",
		},
	}
	for i, j := range names {
		style := discordgo.SecondaryButton
		if i == 0 {
			style = discordgo.SuccessButton
		}
		a = append(a, discordgo.Button{
			Label:    j,
			Style:    style,
			Disabled: true,
			CustomID: fmt.Sprintf("none-%d", i),
		})
	}
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: a},
	}
}

func (b *BigBen) ClickBong(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildId := i.GuildID
	messageId := i.Message.ID
	member := i.Member
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredMessageUpdate})
	b.bongLock.Lock()
	b.currentBong.TriggerClick(guildId, messageId, member.User.ID, member.User.String())
	b.bongLock.Unlock()
}
