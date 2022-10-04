package main

import (
	"fmt"
	"github.com/MrMelon54/BigBen/tables"
	"github.com/MrMelon54/BigBen/utils"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"sync"
	"time"
	"xorm.io/xorm"
)

type CurrentBong struct {
	Engine    *xorm.Engine
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
}

type ClickInfo struct {
	GuildId   snowflake.ID
	MessageId snowflake.ID
	UserId    snowflake.ID
	InterId   snowflake.ID
	Name      string
}

func NewCurrentBong(engine *xorm.Engine, text string, sTime, eTime time.Time) *CurrentBong {
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

func (c *CurrentBong) internalLoop() {
outer:
	for {
		select {
		case <-c.mDone:
			break outer
		case i := <-c.mChan:
			g := c.GuildMapItem(i.GuildId)
			g.Lock.Lock()
			used := false
			won := false
			ct := i.InterId.Time()
			mt := i.MessageId.Time()
			if g.MessageId == i.MessageId {
				for _, j := range g.ClickIds {
					if j == i.UserId {
						goto exitClickCheck
					}
				}
				g.ClickIds = append(g.ClickIds, i.UserId)
				tf := ct.Format("15:04:05.000 UTC")
				ts := ct.Sub(mt)
				g.ClickNames = append(g.ClickNames, fmt.Sprintf("%s | %s | %s", i.Name, tf, ts))
				g.Dirty = true
				used = true

				// click ids should now be 1 if this player won
				if len(g.ClickIds) == 1 {
					won = true
				}
			}
		exitClickCheck:
			g.Lock.Unlock()
			if used {
				_, _ = c.Engine.Insert(&tables.BongLog{
					GuildId: utils.XormSnowflake(i.GuildId),
					UserId:  utils.XormSnowflake(i.UserId),
					MsgId:   utils.XormSnowflake(g.MessageId),
					InterId: utils.XormSnowflake(i.InterId),
					Won:     &won,
				})
			}
		}
	}
}

func (c *CurrentBong) Kill() {
	close(c.mDone)
}

func (c *CurrentBong) GuildMapItem(guildId snowflake.ID) *GuildCurrentBong {
	c.mapLock.RLock()
	g := c.guilds[guildId]
	c.mapLock.RUnlock()
	return g
}

func (c *CurrentBong) RandomGuildData(all []tables.GuildSettings) {
	c.mapLock.Lock()
	for _, i := range all {
		c.guilds[snowflake.ID(i.GuildId)] = &GuildCurrentBong{
			Lock:       &sync.RWMutex{},
			Emoji:      utils.RandomEmoji(i.BongEmoji),
			ClickIds:   []snowflake.ID{},
			ClickNames: []string{},
		}
	}
	c.mapLock.Unlock()
}

func (c *CurrentBong) TriggerClick(event *events.ComponentInteractionCreate) {
	member := event.Member()
	if c.mChan == nil {
		return
	}
	c.mChan <- ClickInfo{
		GuildId:   *event.GuildID(),
		MessageId: event.Message.ID,
		UserId:    member.User.ID,
		InterId:   event.ID(),
		Name:      member.User.Tag(),
	}
}
