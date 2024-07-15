package bigben

import (
	"context"
	"fmt"
	channelSorter "github.com/MrMelon54/channel-sorter"
	"github.com/discord-plays/bigben/database"
	"github.com/discord-plays/bigben/logger"
	"github.com/discord-plays/bigben/utils"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"sync"
	"time"
)

type CurrentBong struct {
	Engine    *database.Queries
	Text      string
	StartTime time.Time
	EndTime   time.Time
	mapLock   *sync.RWMutex
	guilds    map[snowflake.ID]*GuildCurrentBong
	mChan     chan ClickInfo
	mDone     chan struct{}
}

type GuildCurrentBong struct {
	Lock       *sync.RWMutex
	Dirty      bool
	Emoji      string
	MessageId  snowflake.ID
	ClickIds   []snowflake.ID
	ClickNames []string
	In         chan ClickInfo
	T          time.Time // custom start time
}

type ClickInfo struct {
	GuildId       snowflake.ID
	MessageId     snowflake.ID
	UserId        snowflake.ID
	InteractionId snowflake.ID
	Name          string
}

func NewCurrentBong(engine *database.Queries, text string, sTime, eTime time.Time) *CurrentBong {
	c := &CurrentBong{
		Engine:    engine,
		Text:      text,
		StartTime: sTime,
		EndTime:   eTime,
		mapLock:   &sync.RWMutex{},
		guilds:    make(map[snowflake.ID]*GuildCurrentBong),
		mChan:     make(chan ClickInfo, 256),
		mDone:     make(chan struct{}, 1),
	}
	go c.internalLoop()
	return c
}

// internalLoop handles receiving the incoming clicks
//
// TODO(Melon): refactor this
func (c *CurrentBong) internalLoop() {
outer:
	for {
		select {
		case <-c.mDone:
			break outer
		case i := <-c.mChan:
			g := c.GuildMapItem(i.GuildId)
			g.Lock.Lock()
			ct := i.InteractionId.Time()
			mt := i.MessageId.Time()
			ts := ct.Sub(mt)
			if g.MessageId == i.MessageId {
				if ct.Before(c.StartTime) || ct.After(c.EndTime) {
					goto exitClickCheck
				}
				for _, j := range g.ClickIds {
					if j == i.UserId {
						goto exitClickCheck
					}
				}
				g.ClickIds = append(g.ClickIds, i.UserId)
				tf := ct.Format("15:04:05.000 UTC")
				g.ClickNames = append(g.ClickNames, fmt.Sprintf("%s | %s | %s", i.Name, tf, ts))
				g.Dirty = true
				g.In <- i
			}
		exitClickCheck:
			g.Lock.Unlock()
		}
	}
}

// Kill triggers the done state for this CurrentBong
func (c *CurrentBong) Kill() {
	close(c.mDone)
}

// GuildMapItem locks and fetches the GuildCurrentBong for the specified guild
// snowflake
func (c *CurrentBong) GuildMapItem(guildId snowflake.ID) *GuildCurrentBong {
	c.mapLock.RLock()
	g := c.guilds[guildId]
	c.mapLock.RUnlock()
	return g
}

// RandomGuildData generates random GuildCurrentBong structs for each guild using
// the provided settings.
//
// TODO(Melon): refactor this
func (c *CurrentBong) RandomGuildData(all []database.Guild) {
	c.mapLock.Lock()
	for _, i := range all {
		y := &GuildCurrentBong{
			Lock:       &sync.RWMutex{},
			Emoji:      utils.RandomEmoji(i.BongEmoji),
			ClickIds:   []snowflake.ID{},
			ClickNames: []string{},
			In:         make(chan ClickInfo, 10),
		}
		c.guilds[i.ID] = y
		go func() {
			won := true
			z := channelSorter.Sort[ClickInfo](time.Minute*1, y.In, func(a ClickInfo, b ClickInfo) bool {
				return a.InteractionId.Time().Before(b.InteractionId.Time())
			})
		kill:
			for {
				select {
				case <-c.mDone:
					break kill
				case i2 := <-z:
					for _, i := range i2 {
						ct := i.InteractionId.Time()
						mt := i.MessageId.Time()
						ts := ct.Sub(mt)
						n := 0
					tryBongLogInsert:
						err := c.Engine.AddBong(context.Background(), database.AddBongParams{
							GuildID:       i.GuildId,
							UserID:        i.UserId,
							MessageID:     i.MessageId,
							InteractionID: i.InteractionId,
							Won:           won,
							Speed:         ts.Milliseconds(),
						})
						if err != nil {
							if n > 2 {
								logger.Logger.Error("Failed to insert into Bong, giving up", "err", err)
								logger.Logger.Log(logger.DevInsertLevel, "Manual log entry", "row", fmt.Sprintf("%s,%s,%s,%s,%v,%v", i.GuildId, i.UserId, i.MessageId, i.InteractionId, won, ts.Milliseconds()))
							} else {
								logger.Logger.Error("Failed to insert into Bong, trying again", "err", err)
								n++
								goto tryBongLogInsert
							}
						}
						if won {
							won = false
						}
						userId := i.UserId
						tag := i.Name
						// ignoring errors on purpose, I don't remember why?
						err = c.Engine.ReplaceUser(context.Background(), database.ReplaceUserParams{ID: userId, Tag: tag})
						if err != nil {
							logger.Logger.Error("Failed to update user log", "user", userId, "tag", tag, "err", err)
							return
						}
					}
				}
			}
		}()
	}
	c.mapLock.Unlock()
}

func (c *CurrentBong) TriggerClick(event *events.ComponentInteractionCreate) {
	member := event.Member()
	if c.mChan == nil {
		return
	}
	c.mChan <- ClickInfo{
		GuildId:       *event.GuildID(),
		MessageId:     event.Message.ID,
		UserId:        member.User.ID,
		InteractionId: event.ID(),
		Name:          member.User.Tag(),
	}
}
