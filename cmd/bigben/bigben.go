package main

import (
	"context"
	"fmt"
	"github.com/MrMelon54/BigBen/assets"
	"github.com/MrMelon54/BigBen/commands"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"xorm.io/xorm"
)

const (
	bongCron          = "0 0 * * * *"   // @hourly
	bongSetupCron     = "0 50 * * * *"  // @hourly at 50min
	bongDebugCron     = "0 */5 * * * *" // every 5min
	updateMessageCron = "*/2 * * * * *" // every 2sec
)

var intents = []gateway.Intents{
	gateway.IntentGuilds,
	gateway.IntentGuildMessages,
}

type BigBen struct {
	engine          *xorm.Engine
	appId           snowflake.ID
	guildId         snowflake.ID
	client          bot.Client
	commands        commands.CommandList
	commandHandlers map[string]commands.CommandHandler
	bongLock        *sync.Mutex
	currentBong     *CurrentBong
	cron            *cron.Cron
}

func (b *BigBen) Engine() *xorm.Engine {
	return b.engine
}

func (b *BigBen) AppId() snowflake.ID   { return b.appId }
func (b *BigBen) GuildId() snowflake.ID { return b.guildId }
func (b *BigBen) Session() bot.Client   { return b.client }

func NewBigBen(engine *xorm.Engine, token string, appId, guildId snowflake.ID) (*BigBen, error) {
	client, err := disgo.New(token, bot.WithCacheConfigOpts(
		cache.WithCacheFlags(cache.FlagVoiceStates, cache.FlagMembers, cache.FlagChannels, cache.FlagGuilds, cache.FlagRoles),
	), bot.WithGatewayConfigOpts(
		gateway.WithIntents(intents...),
		gateway.WithCompress(true),
	))
	if err != nil {
		return nil, err
	}
	client.AddEventListeners(&events.ListenerAdapter{OnReady: func(event *events.Ready) {
		log.Printf("[Ready] Starting BigBen as %s\n", event.User.Tag())
		_ = client.SetPresence(context.Background(), func(presenceUpdate *gateway.MessageDataPresenceUpdate) {
			presenceUpdate.Activities = []discord.Activity{
				{
					Name: "bong",
					Type: discord.ActivityTypeListening,
				},
			}
			presenceUpdate.Status = discord.OnlineStatusOnline
		})
	}})
	return (&BigBen{}).init(engine, appId, guildId, client)
}
func (b *BigBen) init(engine *xorm.Engine, appId, guildId snowflake.ID, client bot.Client) (*BigBen, error) {
	b.engine = engine
	b.appId = appId
	b.guildId = guildId
	b.client = client
	b.commands, b.commandHandlers = commands.InitCommands(b)
	b.bongLock = &sync.Mutex{}
	b.currentBong = nil
	err := b.updateCommands()
	if err != nil {
		return nil, err
	}
	client.AddEventListeners(bot.NewListenerFunc[*events.ApplicationCommandInteractionCreate](func(event *events.ApplicationCommandInteractionCreate) {
		if h, ok := b.commandHandlers[event.Data.CommandName()]; ok {
			h.Handler(event)
		}
	}), bot.NewListenerFunc[*events.ComponentInteractionCreate](func(event *events.ComponentInteractionCreate) {
		if event.Data.Type() == discord.ComponentTypeButton && event.Data.CustomID() == "bong" {
			b.ClickBong(event)
		}
	}))
	b.bingSetup()
	b.cron = cron.New(cron.WithSeconds())
	if os.Getenv("DEBUG_MODE") == "1" {
		_, _ = b.cron.AddFunc(bongDebugCron, b.bingBong)
	} else {
		_, _ = b.cron.AddFunc(bongCron, b.bingBong)
	}
	_, _ = b.cron.AddFunc(updateMessageCron, b.updateMessageData)
	_, _ = b.cron.AddFunc(bongSetupCron, b.bingSetup)
	b.cron.Start()
	return b, nil
}

func (b *BigBen) RunAndBlock() {
	if err := b.client.OpenGateway(context.TODO()); err != nil {
		log.Printf("Failed to connect to gateway: %s", err)
	}

	log.Println("[Main] BigBen is now bonging. Press CTRL-C for maintenance.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	b.Exit()
}

func (b *BigBen) Exit() {
	b.client.Close(context.TODO())
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
	b.currentBong = NewCurrentBong(b.engine, title.S, sTime, eTime)
	b.currentBong.RandomGuildData(all)
	wg := &sync.WaitGroup{}
	for _, i := range all {
		g := b.currentBong.GuildMapItem(snowflake.ID(i.GuildId))
		wg.Add(1)
		go b.internalSendBongMessage(wg, g, i, b.currentBong.Text, utils.ConvertToComponentEmoji(g.Emoji), sTime)
	}
	wg.Wait()
	b.bongLock.Unlock()
}

func (b *BigBen) bingSetup() {
	all, err := b.GetAllGuildSettings()
	if err != nil {
		log.Printf("[bingBong()] Error: %s\n", err)
		return
	}
	wg := &sync.WaitGroup{}
	n := utils.GetStartOfHourTime().Add(time.Hour)
	icon := *assets.ReadClockFaceByTimeAsOptionalIcon(n)
	for _, i := range all {
		wg.Add(1)
		go b.internalSetupWebhook(wg, i, icon)
	}
	wg.Wait()
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
		g := b.currentBong.GuildMapItem(snowflake.ID(i.GuildId))
		if g == nil {
			continue
		}
		g.Lock.Lock()
		if g.Dirty {
			g.Dirty = false
			go b.internalEditBongMessage(i, g.MessageId, b.currentBong.Text, utils.ConvertToComponentEmoji(g.Emoji), g.ClickNames, b.currentBong.StartTime)
			go b.internalBongRoleAssign(i, g.MessageId, g.ClickIds)
		}
		g.Lock.Unlock()
	}
	b.bongLock.Unlock()
}

func (b *BigBen) updateCommands() error {
	_, err := b.client.Rest().SetGlobalCommands(b.appId, b.commands)
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

func (b *BigBen) GetGuildSettings(guildId snowflake.ID) (tables.GuildSettings, error) {
	var g tables.GuildSettings
	_, err := b.engine.Where("guild_id = ?", guildId.String()).Get(&g)
	g.GuildId = utils.XormSnowflake(guildId)
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

func (b *BigBen) internalSetupWebhook(wg *sync.WaitGroup, conf tables.GuildSettings, icon discord.Icon) {
	defer wg.Done()
	_, _ = b.client.Rest().UpdateWebhook(snowflake.ID(conf.BongWebhookId), discord.WebhookUpdate{
		Avatar: json.NewOptional[discord.Icon](icon),
	})
}

func (b *BigBen) internalSendBongMessage(wg *sync.WaitGroup, g *GuildCurrentBong, conf tables.GuildSettings, title string, emoji discord.ComponentEmoji, startTime time.Time) {
	defer wg.Done()
	builder := discord.NewWebhookMessageCreateBuilder()
	builder.SetEmbeds(b.bongEmbeds(title, startTime))
	builder.SetContainerComponents(b.bongComponents(emoji, []string{}))
	m, err := b.client.Rest().CreateWebhookMessage(snowflake.ID(conf.BongWebhookId), conf.BongWebhookToken, builder.Build(), true, 0)
	if err != nil {
		log.Printf("[internalSendBongMessage(\"%s/%s\")] Error: %s\n", snowflake.ID(conf.GuildId), snowflake.ID(conf.BongChannelId), err)
		return
	}
	g.Lock.Lock()
	g.MessageId = m.ID
	g.Lock.Unlock()
}

func (b *BigBen) internalEditBongMessage(conf tables.GuildSettings, messageId snowflake.ID, title string, emoji discord.ComponentEmoji, names []string, startTime time.Time) {
	if messageId == 0 {
		return
	}
	builder := discord.NewWebhookMessageUpdateBuilder()
	builder.SetEmbeds(b.bongEmbeds(title, startTime))
	builder.SetContainerComponents(b.bongComponents(emoji, names))
	_, err := b.client.Rest().UpdateWebhookMessage(snowflake.ID(conf.BongWebhookId), conf.BongWebhookToken, messageId, builder.Build(), 0)
	if err != nil {
		log.Printf("[internalEditBongMessage(\"%s/%s\")] Error: %s\n", snowflake.ID(conf.GuildId), snowflake.ID(conf.BongChannelId), err)
	}
}

func (b *BigBen) internalBongRoleAssign(conf tables.GuildSettings, messageId snowflake.ID, clickIds []snowflake.ID) {
	if conf.BongRoleId == 0 {
		return
	}
	if len(clickIds) < 1 {
		return
	}
	var a []tables.RoleLog
	err := b.engine.Where("guild_id = ? and message_id != ?", conf.GuildId, utils.XormSnowflake(messageId)).Find(&a)
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Database error (get role log row): %s\n", err)
		return
	}
	c := make([]int64, len(a))
	for i, row := range a {
		c[i] = row.Id
		if snowflake.ID(row.UserId) == clickIds[0] {
			continue
		}
		err = b.client.Rest().RemoveMemberRole(snowflake.ID(row.GuildId), snowflake.ID(row.UserId), snowflake.ID(row.RoleId))
		if err != nil {
			log.Printf("[internalBongRoleAssign()] Failed to remove guild member role: %s\n", err)
		}
	}
	_, err = b.engine.In("id", c).Delete(&tables.RoleLog{})
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Database error (delete checked ids): %s\n", err)
	}

	// Just assign the role and let Discord check it
	_, err = b.engine.Insert(&tables.RoleLog{
		GuildId:   conf.GuildId,
		MessageId: utils.XormSnowflake(messageId),
		RoleId:    conf.BongRoleId,
		UserId:    utils.XormSnowflake(clickIds[0]),
	})
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Database error (insert role log row): %s\n", err)
	}
	err = b.client.Rest().AddMemberRole(snowflake.ID(conf.GuildId), clickIds[0], snowflake.ID(conf.BongRoleId))
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Failed to add guild member: %s\n", err)
	}
}

func (b *BigBen) bongEmbeds(title string, t time.Time) discord.Embed {
	return discord.Embed{Title: title, Color: 0xd4af37, Timestamp: &t}
}

func (b *BigBen) bongComponents(emoji discord.ComponentEmoji, names []string) discord.ContainerComponent {
	if len(names) > 3 {
		names = names[:3]
	}
	l := ""
	if len(names) == 1 {
		l = "1 click"
	} else if len(names) > 1 {
		l = fmt.Sprintf("%d clicks", len(names))
	}
	a := []discord.InteractiveComponent{discord.NewSecondaryButton(l, "bong").WithEmoji(emoji)}
	for i, j := range names {
		style := discord.ButtonStyleSecondary
		if i == 0 {
			style = discord.ButtonStyleSuccess
		}
		a = append(a, discord.NewButton(discord.ButtonStyle(style), j, fmt.Sprintf("none-%d", i), "").AsDisabled())
	}
	return discord.NewActionRow(a...)
}

func (b *BigBen) ClickBong(event *events.ComponentInteractionCreate) {
	_ = event.Respond(discord.InteractionResponseTypeDeferredUpdateMessage, nil)
	b.bongLock.Lock()
	if b.currentBong != nil {
		b.currentBong.TriggerClick(event)
	}
	b.bongLock.Unlock()
}
