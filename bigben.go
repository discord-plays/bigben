package bigben

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/discord-plays/bigben/assets"
	"github.com/discord-plays/bigben/commands"
	"github.com/discord-plays/bigben/database"
	"github.com/discord-plays/bigben/inter"
	"github.com/discord-plays/bigben/message"
	"github.com/discord-plays/bigben/utils"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/robfig/cron/v3"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	bongCron          = "0 0 * * * *"    // @hourly
	bongStatusUpdate  = "0 * * * * *"    // @minutely
	bongSetupCron     = "0 50 * * * *"   // @hourly at 50min
	bongDebugCron     = "0 */2 * * * *"  // every 2min
	updateMessageCron = "*/2 * * * * *"  // every 2sec
	bongNewYearCron   = "0 10 0 1 1 *"   // New Years day at 0:10
	bongChristmasCron = "0 10 0 25 12 *" // Christmas day at 0:10
)

var intents = []gateway.Intents{
	gateway.IntentGuilds,
	gateway.IntentGuildMessages,
}

// BigBen contains all the commands, config and cron trigger logic
type BigBen struct {
	engine             *database.Queries
	appId              snowflake.ID
	guildId            snowflake.ID
	client             bot.Client
	uploadToken        string
	commands           commands.CommandList
	commandHandlers    map[string]commands.CommandHandler
	bongLock           *sync.Mutex
	oldBong            *CurrentBong
	currentBong        *CurrentBong
	cron               *cron.Cron
	statusPushEndpoint string
}

var _ inter.MainBotInterface = &BigBen{}

func (b *BigBen) Engine() *database.Queries { return b.engine }
func (b *BigBen) AppId() snowflake.ID       { return b.appId }
func (b *BigBen) GuildId() snowflake.ID     { return b.guildId }
func (b *BigBen) Session() bot.Client       { return b.client }

// NewBigBen creates a new instance of the BigBen struct
func NewBigBen(engine *database.Queries, token, uploadToken, statusPush string, appId, guildId snowflake.ID) (*BigBen, error) {
	client, err := disgo.New(token, bot.WithCacheConfigOpts(
		cache.WithCaches(cache.FlagVoiceStates, cache.FlagMembers, cache.FlagChannels, cache.FlagGuilds, cache.FlagRoles),
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
	return (&BigBen{}).init(engine, appId, guildId, client, uploadToken, statusPush)
}

func (b *BigBen) init(engine *database.Queries, appId, guildId snowflake.ID, client bot.Client, uploadToken, statusPush string) (*BigBen, error) {
	// fill parameters
	b.engine = engine
	b.appId = appId
	b.guildId = guildId
	b.client = client
	b.uploadToken = uploadToken
	b.commands, b.commandHandlers = commands.InitCommands(b)
	b.bongLock = &sync.Mutex{}
	b.currentBong = nil
	b.statusPushEndpoint = statusPush

	// try to update commands and add event listeners
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

	// setup for the next bong
	b.bingSetup()

	// setup cron library with second support
	b.cron = cron.New(cron.WithSeconds())

	// debug mode sends a bong every 2 minutes
	if os.Getenv("DEBUG_MODE") == "1" {
		_, _ = b.cron.AddFunc(bongDebugCron, b.bingBong)
	} else {
		_, _ = b.cron.AddFunc(bongCron, b.bingBong)
	}

	// setup status push task to call every minute
	if b.statusPushEndpoint != "" {
		_, _ = b.cron.AddFunc(bongStatusUpdate, b.statusUpdate)
	}

	// setup Christmas notifications
	cronChristmas := b.messageNotification("Christmas", message.SendChristmasNotification)

	// setup update message, bong setup, Christmas and New Year tasks
	_, _ = b.cron.AddFunc(updateMessageCron, b.updateMessageData)
	_, _ = b.cron.AddFunc(bongSetupCron, b.bingSetup)
	_, _ = b.cron.AddFunc(bongChristmasCron, cronChristmas)
	_, _ = b.cron.AddFunc(bongNewYearCron, b.cronNewYears)

	// start the cron scheduler
	b.cron.Start()

	// map of debug commands for testing calls
	commands.DebugCommands = map[string]func(){
		"bingBong":  b.bingBong,
		"bingSetup": b.bingSetup,
		"christmas": cronChristmas,
		"newYears":  b.cronNewYears,
	}
	return b, nil
}

// RunAndBlock connects to the Discord gateway and starts the main bot sequence.
// This method blocks until sent an interrupt signal.
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

// Exit closes the Discord client connection
func (b *BigBen) Exit() {
	b.client.Close(context.TODO())
}

// statusUpdate send a push message to the status endpoint
func (b *BigBen) statusUpdate() {
	v := url.Values{
		"status": []string{"up"},
		"msg":    []string{"OK"},
		"ping":   []string{fmt.Sprintf("%d", b.client.Gateway().Latency().Milliseconds())},
	}
	get, err := http.Get(b.statusPushEndpoint + "?" + v.Encode())
	if err != nil {
		return
	}
	_ = get.Body.Close()
}

// bingBong sends the hourly bong message to all guilds
func (b *BigBen) bingBong() {
	log.Println("[bingBong()] Sending hourly bong")

	// read guild settings
	all, err := b.engine.GetAllGuilds(context.Background())
	if err != nil {
		log.Printf("[bingBong()] Error: %s\n", err)
		return
	}

	// calculate the start and end time
	now := utils.GetStartOfHourTime()
	title := utils.GetBongTitle(now)
	sTime := now
	eTime := now.Add(time.Hour)

	// generate a new bong with random guild data
	currentBong := NewCurrentBong(b.engine, title.S, sTime, eTime)
	currentBong.RandomGuildData(all)

	// lock while generating and swapping to new bong
	b.bongLock.Lock()

	// kill the old bong if it's still running
	if b.oldBong != nil {
		b.oldBong.Kill()
	}

	// move the current bong to be the old bong
	b.currentBong = currentBong
	b.oldBong = b.currentBong

	// finish with the lock
	b.bongLock.Unlock()

	// send bong message to all guilds
	for _, i := range all {
		g := currentBong.GuildMapItem(i.ID)
		g.T = sTime
		go b.internalSendBongMessage(g, i, currentBong.Text, utils.ConvertToComponentEmoji(g.Emoji), sTime, title.A)
	}
}

// bingSetup changes the webhook profile picture each hour
func (b *BigBen) bingSetup() {
	// load guild settings
	all, err := b.engine.GetAllGuilds(context.Background())
	if err != nil {
		log.Printf("[bingBong()] Error: %s\n", err)
		return
	}

	// setup wait group to wait for all actions to finish
	wg := &sync.WaitGroup{}

	// get the current hour and find the clock face icon
	n := utils.GetStartOfHourTime().Add(time.Hour)
	icon := *assets.ReadClockFaceByTimeAsOptionalIcon(n)

	// modify the webhook in each guild
	for _, i := range all {
		wg.Add(1)
		go b.internalSetupWebhook(wg, i, icon)
	}
	wg.Wait()
}

// updateMessageData edits messages and the bong role
func (b *BigBen) updateMessageData() {
	// don't try and update if currentBong is not initialised yet
	if b.currentBong == nil {
		return
	}

	// load guild settings
	all, err := b.engine.GetAllGuilds(context.Background())
	if err != nil {
		log.Printf("[updateMessageData()] Error: %s\n", err)
		return
	}

	// get copies of the old and current bongs
	b.bongLock.Lock()
	oldBong := b.oldBong
	currentBong := b.currentBong
	b.bongLock.Unlock()

	// loop over all guilds and update bong data for the oldBong and currentBong
	for _, i := range all {
		if oldBong != nil {
			b.updateBongData(i, oldBong)
		}

		// update current bong data
		b.updateBongData(i, currentBong)
	}
}

func (b *BigBen) updateBongData(i database.Guild, bong *CurrentBong) {
	// find the guild in the current bong map
	g := bong.GuildMapItem(i.ID)
	if g == nil {
		return
	}

	// lock while grabbing parameters
	g.Lock.Lock()
	if g.Dirty {
		g.Dirty = false

		// launch edit bong message and bong role assign in goroutines
		go b.internalEditBongMessage(i, g.MessageId, bong.Text, utils.ConvertToComponentEmoji(g.Emoji), g.ClickNames, g.T)
		go b.internalBongRoleAssign(i, g.MessageId, g.ClickIds)
	}
	g.Lock.Unlock()
}

// updateCommands sets up the global commands for the bot
func (b *BigBen) updateCommands() error {
	// if a guildId is set then the global commands are only updated for that guild
	if b.guildId == 0 {
		_, err := b.client.Rest().SetGlobalCommands(b.appId, b.commands)
		if err != nil {
			return fmt.Errorf("bulk overwrite global application commands error: %s", err)
		}
	} else {
		_, err := b.client.Rest().SetGuildCommands(b.appId, b.guildId, b.commands)
		if err != nil {
			return fmt.Errorf("bulk overwrite guild application commands error: %s", err)
		}
	}
	return nil
}

// internalSetupWebhook updates the avatar for the webhook by ID
func (b *BigBen) internalSetupWebhook(wg *sync.WaitGroup, conf database.Guild, icon discord.Icon) {
	defer wg.Done()
	_, _ = b.client.Rest().UpdateWebhook(conf.BongWebhookID, discord.WebhookUpdate{
		Avatar: json.NewNullablePtr[discord.Icon](icon),
	})
}

// internalSendBongMessage sends a bong message for the specified guild
func (b *BigBen) internalSendBongMessage(g *GuildCurrentBong, conf database.Guild, title string, emoji discord.ComponentEmoji, startTime time.Time, aprilFools bool) {
	// wait for a random number of minutes on April fools
	if aprilFools {
		waitMin := time.Minute * time.Duration(rand.Intn(30))
		startTime = startTime.Add(waitMin)
		g.T = startTime
		if os.Getenv("DEBUG_MODE") == "1" {
			fmt.Printf("[Debug] Delaying %s for %s\n", conf.ID, waitMin)
		}
		<-time.After(time.Until(startTime))
	}

	// build webhook message
	builder := discord.NewWebhookMessageCreateBuilder()
	builder.SetEmbeds(b.bongEmbeds(title, startTime))
	builder.SetContainerComponents(b.bongComponents(emoji, []string{}))

	// send webhook message
	m, err := b.client.Rest().CreateWebhookMessage(conf.BongWebhookID, conf.BongWebhookToken, builder.Build(), true, 0)
	if err != nil {
		log.Printf("[internalSendBongMessage(\"%s/%s\")] Error: %s\n", conf.ID, conf.BongChannelID, err)
		return
	}

	// lock and update message ID
	g.Lock.Lock()
	g.MessageId = m.ID
	g.Lock.Unlock()
}

// internalEditBongMessage changes the contents of the bong message
func (b *BigBen) internalEditBongMessage(conf database.Guild, messageId snowflake.ID, title string, emoji discord.ComponentEmoji, names []string, startTime time.Time) {
	if messageId == 0 {
		return
	}

	// build webhook message updater
	builder := discord.NewWebhookMessageUpdateBuilder()
	builder.SetEmbeds(b.bongEmbeds(title, startTime))
	builder.SetContainerComponents(b.bongComponents(emoji, names))

	// update webhook message
	_, err := b.client.Rest().UpdateWebhookMessage(conf.BongWebhookID, conf.BongWebhookToken, messageId, builder.Build(), 0)
	if err != nil {
		log.Printf("[internalEditBongMessage(\"%s/%s\")] Error: %s\n", conf.ID, conf.BongChannelID, err)
	}
}

// internalBongRoleAssign removes the bong role from the members who currently have it and adds the bong role to the winning member
func (b *BigBen) internalBongRoleAssign(conf database.Guild, messageId snowflake.ID, clickIds []snowflake.ID) {
	if conf.BongRoleID == 0 {
		return
	}
	if len(clickIds) < 1 {
		return
	}

	// load role logs
	roleLogs, err := b.engine.GetRoleLogs(context.Background(), database.GetRoleLogsParams{GuildID: conf.ID, MessageID: messageId})
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Database error (get role log row): %s\n", err)
		return
	}

	// remove bong role from all members except the winning member
	c := make([]int64, len(roleLogs))
	for i, row := range roleLogs {
		c[i] = row.ID
		// if the UserID is the winning member then continue and do not remove the bong role
		if row.UserID == clickIds[0] {
			continue
		}
		err = b.client.Rest().RemoveMemberRole(row.GuildID, row.UserID, row.RoleID)
		if err != nil {
			log.Printf("[internalBongRoleAssign()] Failed to remove guild member role: %s\n", err)
		}
	}

	// delete all the rows from the previously fetched role log
	err = b.engine.DeleteRoles(context.Background(), c)
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Database error (delete checked ids): %s\n", err)
	}

	// Just assign the role and let Discord check it
	err = b.engine.AddRole(context.Background(), database.AddRoleParams{
		GuildID:   conf.ID,
		MessageID: messageId,
		RoleID:    conf.BongRoleID,
		UserID:    clickIds[0],
	})
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Database error (insert role log row): %s\n", err)
	}

	// add the role to the winning member
	err = b.client.Rest().AddMemberRole(conf.ID, clickIds[0], conf.BongRoleID)
	if err != nil {
		log.Printf("[internalBongRoleAssign()] Failed to add guild member: %s\n", err)
	}
}

// bongEmbeds returns the Discord embed with the title and timestamp
func (b *BigBen) bongEmbeds(title string, t time.Time) discord.Embed {
	return discord.Embed{Title: title, Color: 0xd4af37, Timestamp: &t}
}

// bongComponents returns the button components below the bong message
func (b *BigBen) bongComponents(emoji discord.ComponentEmoji, names []string) discord.ContainerComponent {
	// limit to displaying the top 3 members
	if len(names) > 3 {
		names = names[:3]
	}
	// format click/clicks text
	l := ""
	if len(names) == 1 {
		l = "1 click"
	} else if len(names) > 1 {
		l = fmt.Sprintf("%d clicks", len(names))
	}

	// setup interactive component slice with bong button
	rowButtons := []discord.InteractiveComponent{discord.NewSecondaryButton(l, "bong").WithEmoji(emoji)}

	// loop over winning names and add them in disabled buttons
	// the 0th button has a "success" style applied to it
	for i, j := range names {
		style := discord.ButtonStyleSecondary
		if i == 0 {
			style = discord.ButtonStyleSuccess
		}
		rowButtons = append(rowButtons, discord.NewButton(discord.ButtonStyle(style), j, fmt.Sprintf("none-%d", i), "").AsDisabled())
	}

	if len(names) > 3 {
		rowButtons = append(rowButtons, discord.NewButton(discord.ButtonStyleSecondary, fmt.Sprintf("+%d", len(names)-3), "more", "").AsDisabled())
	}

	// return the buttons in an action row
	return discord.NewActionRow(rowButtons...)
}

// ClickBong is triggered on interaction events
func (b *BigBen) ClickBong(event *events.ComponentInteractionCreate) {
	// respond with a deferred update to render the interaction as finished
	_ = event.Respond(discord.InteractionResponseTypeDeferredUpdateMessage, nil)

	// lock and trigger the old and current bongs
	b.bongLock.Lock()
	if b.oldBong != nil {
		b.oldBong.TriggerClick(event)
	}
	if b.currentBong != nil {
		b.currentBong.TriggerClick(event)
	}
	b.bongLock.Unlock()
}
